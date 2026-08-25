import { createHash } from "node:crypto";
import Stripe from "stripe";
import { createAdminClient } from "@/lib/supabase/admin";
import { sendEmail } from "@/lib/email/resend";
import { getSiteSettings } from "@/features/settings/queries";
import {
  findContributionByInvoice,
  findContributionByPaymentIntent,
  findRecurringBySubscription,
} from "@/features/payments/services";
import { formatCurrency } from "@/lib/format";
import type { RecurringStatus } from "@/lib/db-types";

const admin = () => createAdminClient();

export async function handleStripeWebhook(event: Stripe.Event): Promise<{
  status: "processed" | "skipped";
}> {
  const hash = createHash("sha256").update(event.id).digest("hex");

  const { data: existing } = await admin()
    .from("stripe_webhook_events")
    .select("id")
    .eq("stripe_event_id", event.id)
    .maybeSingle();

  if (existing) {
    return { status: "skipped" };
  }

  try {
    switch (event.type) {
      case "payment_intent.succeeded":
        await handlePaymentIntentSucceeded(event.data.object);
        break;
      case "payment_intent.payment_failed":
        await handlePaymentIntentFailed(event.data.object);
        break;
      case "payment_intent.canceled":
        await handlePaymentIntentCanceled(event.data.object);
        break;
      case "charge.refunded":
        await handleChargeRefunded(event.data.object);
        break;
      case "invoice.payment_succeeded":
        await handleInvoicePaymentSucceeded(event.data.object);
        break;
      case "invoice.payment_failed":
        await handleInvoicePaymentFailed(event.data.object);
        break;
      case "customer.subscription.created":
      case "customer.subscription.updated":
        await syncSubscription(event.data.object);
        break;
      case "customer.subscription.deleted":
        await handleSubscriptionDeleted(event.data.object);
        break;
      default:
        break;
    }

    await admin().from("stripe_webhook_events").insert({
      stripe_event_id: event.id,
      event_type: event.type,
      payload_hash: hash,
      processing_status: "processed",
    });
    return { status: "processed" };
  } catch (error) {
    await admin().from("stripe_webhook_events").insert({
      stripe_event_id: event.id,
      event_type: event.type,
      payload_hash: hash,
      processing_status: "failed",
      error_message: error instanceof Error ? error.message : "Erro desconhecido",
    });
    throw error;
  }
}

function metadataOf(obj: {
  metadata?: Stripe.Metadata | undefined;
}): {
  donor_id?: string;
  community_id?: string;
  community_name?: string;
  payment_type?: string;
} {
  return {
    donor_id: obj.metadata?.donor_id,
    community_id: obj.metadata?.community_id || undefined,
    community_name: obj.metadata?.community_name || undefined,
    payment_type: obj.metadata?.payment_type,
  };
}

async function ensureContributionRecord(input: {
  donorId: string;
  communityId: string | null;
  communityName?: string | null;
  recurringDonationId?: string | null;
  amountCents: number;
  paymentType: "one_time" | "recurring";
  paymentIntentId?: string | null;
  invoiceId?: string | null;
  subscriptionId?: string | null;
  eventId?: string | null;
  status: "paid" | "failed" | "canceled" | "refunded" | "pending";
  paidAt?: string | null;
  refundedAt?: string | null;
}) {
  const paymentIntentId = input.paymentIntentId ?? null;
  const invoiceId = input.invoiceId ?? null;
  const existing = paymentIntentId
    ? await findContributionByPaymentIntent(paymentIntentId)
    : invoiceId
      ? await findContributionByInvoice(invoiceId)
      : null;

  const payload = {
    donor_id: input.donorId,
    community_id: input.communityId,
    community_name: input.communityName ?? null,
    recurring_donation_id: input.recurringDonationId ?? null,
    amount_cents: input.amountCents,
    currency: "brl",
    payment_type: input.paymentType,
    status: input.status,
    stripe_payment_intent_id: paymentIntentId,
    stripe_invoice_id: invoiceId,
    stripe_subscription_id: input.subscriptionId ?? null,
    stripe_event_id: input.eventId ?? null,
    paid_at: input.paidAt ?? null,
    refunded_at: input.refundedAt ?? null,
  };

  if (existing) {
    await admin()
      .from("contribution_payments")
      .update(payload)
      .eq("id", existing.id);
  } else {
    await admin().from("contribution_payments").insert(payload);
  }
}

async function handlePaymentIntentSucceeded(pi: Stripe.PaymentIntent) {
  const meta = metadataOf(pi);
  if (!meta.donor_id) return;
  await ensureContributionRecord({
    donorId: meta.donor_id,
    communityId: meta.community_id ?? null,
    communityName: meta.community_name ?? null,
    amountCents: pi.amount,
    paymentType: meta.payment_type === "recurring" ? "recurring" : "one_time",
    paymentIntentId: pi.id,
    eventId: pi.id,
    status: "paid",
    paidAt: new Date().toISOString(),
  });
}

async function handlePaymentIntentFailed(pi: Stripe.PaymentIntent) {
  const meta = metadataOf(pi);
  if (!meta.donor_id) return;
  await ensureContributionRecord({
    donorId: meta.donor_id,
    communityId: meta.community_id ?? null,
    communityName: meta.community_name ?? null,
    amountCents: pi.amount,
    paymentType: meta.payment_type === "recurring" ? "recurring" : "one_time",
    paymentIntentId: pi.id,
    status: "failed",
  });
}

async function handlePaymentIntentCanceled(pi: Stripe.PaymentIntent) {
  const meta = metadataOf(pi);
  if (!meta.donor_id) return;
  await ensureContributionRecord({
    donorId: meta.donor_id,
    communityId: meta.community_id ?? null,
    communityName: meta.community_name ?? null,
    amountCents: pi.amount,
    paymentType: meta.payment_type === "recurring" ? "recurring" : "one_time",
    paymentIntentId: pi.id,
    status: "canceled",
  });
}

async function handleChargeRefunded(charge: Stripe.Charge) {
  if (!charge.payment_intent) return;
  const existing = await findContributionByPaymentIntent(
    typeof charge.payment_intent === "string"
      ? charge.payment_intent
      : charge.payment_intent.id
  );
  if (!existing) return;
  await admin()
    .from("contribution_payments")
    .update({
      status: "refunded",
      refunded_at: new Date().toISOString(),
    })
    .eq("id", existing.id);
}

async function resolveSubscriptionInfo(
  invoice: Stripe.Invoice
): Promise<{
  donorId: string | null;
  communityId: string | null;
  communityName: string | null;
  recurringId: string | null;
  subscriptionId: string | null;
}> {
  const subscriptionDetails = invoice.parent?.subscription_details;
  const rawSubscription = subscriptionDetails?.subscription;
  const subscriptionId =
    typeof rawSubscription === "string" ? rawSubscription : rawSubscription?.id ?? null;

  if (subscriptionId) {
    const local = await findRecurringBySubscription(subscriptionId);
    if (local) {
      const { data: donor } = await admin()
        .from("donors")
        .select("community_name")
        .eq("id", local.donor_id)
        .maybeSingle();
      return {
        donorId: local.donor_id,
        communityId: local.community_id,
        communityName: donor?.community_name ?? null,
        recurringId: local.id,
        subscriptionId,
      };
    }
  }

  const donorId = subscriptionDetails?.metadata?.donor_id ?? null;
  const communityId = subscriptionDetails?.metadata?.community_id || null;
  const communityName =
    subscriptionDetails?.metadata?.community_name || null;
  return {
    donorId,
    communityId,
    communityName,
    recurringId: null,
    subscriptionId,
  };
}

async function handleInvoicePaymentSucceeded(invoice: Stripe.Invoice) {
  if (invoice.status !== "paid") return;
  const { donorId, communityId, communityName, recurringId, subscriptionId } =
    await resolveSubscriptionInfo(invoice);
  if (!donorId) return;

  const amount = invoice.amount_paid ?? invoice.amount_due ?? 0;
  await ensureContributionRecord({
    donorId,
    communityId,
    communityName,
    recurringDonationId: recurringId,
    amountCents: amount,
    paymentType: "recurring",
    invoiceId: invoice.id,
    subscriptionId,
    eventId: invoice.id,
    status: "paid",
    paidAt: new Date().toISOString(),
  });

  if (recurringId) {
    await admin()
      .from("recurring_donations")
      .update({ status: "active" })
      .eq("id", recurringId);
  }
  await notifyReceipt(invoice, donorId, amount);
}

async function handleInvoicePaymentFailed(invoice: Stripe.Invoice) {
  const { donorId, communityId, communityName, recurringId, subscriptionId } =
    await resolveSubscriptionInfo(invoice);
  if (!donorId) return;

  const amount = invoice.amount_due ?? 0;
  await ensureContributionRecord({
    donorId,
    communityId,
    communityName,
    recurringDonationId: recurringId,
    amountCents: amount,
    paymentType: "recurring",
    invoiceId: invoice.id,
    subscriptionId,
    status: "failed",
  });

  if (recurringId) {
    await admin()
      .from("recurring_donations")
      .update({ status: "past_due" })
      .eq("id", recurringId);
  }
}

function mapSubscriptionStatus(
  sub: Stripe.Subscription
): { status: RecurringStatus; canceledAt: string | null; cancelAtPeriodEnd: boolean; currentPeriodEnd: string | null } {
  const currentPeriodEnd = sub.items?.data?.[0]?.current_period_end
    ? new Date(sub.items.data[0].current_period_end * 1000).toISOString()
    : null;
  const cancelAtPeriodEnd = sub.cancel_at_period_end === true;
  const canceledAt =
    sub.status === "canceled" || sub.canceled_at
      ? new Date((sub.canceled_at ?? Date.now() / 1000) * 1000).toISOString()
      : null;

  if (sub.status === "canceled") {
    return { status: "canceled", canceledAt, cancelAtPeriodEnd: false, currentPeriodEnd };
  }
  if (cancelAtPeriodEnd) {
    return { status: "cancel_at_period_end", canceledAt: null, cancelAtPeriodEnd: true, currentPeriodEnd };
  }
  if (sub.status === "past_due" || sub.status === "incomplete" || sub.status === "incomplete_expired") {
    return { status: "past_due", canceledAt: null, cancelAtPeriodEnd: false, currentPeriodEnd };
  }
  return { status: "active", canceledAt: null, cancelAtPeriodEnd: false, currentPeriodEnd };
}

async function syncSubscription(sub: Stripe.Subscription) {
  const local = await findRecurringBySubscription(sub.id);
  if (!local) return;
  const mapped = mapSubscriptionStatus(sub);
  await admin()
    .from("recurring_donations")
    .update({
      status: mapped.status,
      cancel_at_period_end: mapped.cancelAtPeriodEnd,
      canceled_at: mapped.canceledAt,
      current_period_end: mapped.currentPeriodEnd,
    })
    .eq("id", local.id);
}

async function handleSubscriptionDeleted(sub: Stripe.Subscription) {
  const local = await findRecurringBySubscription(sub.id);
  if (!local) return;
  await admin()
    .from("recurring_donations")
    .update({
      status: "canceled",
      cancel_at_period_end: false,
      canceled_at: new Date().toISOString(),
    })
    .eq("id", local.id);
}

async function notifyReceipt(
  invoice: Stripe.Invoice,
  donorId: string,
  amountCents: number
) {
  const { data: donor } = await admin()
    .from("donors")
    .select("email, full_name")
    .eq("id", donorId)
    .single();

  if (!donor?.email) return;

  const settings = await getSiteSettings();

  const html = `
    <div style="font-family: Arial, sans-serif; max-width: 600px;">
      <h2>Recibo da sua contribuição</h2>
      <p>Olá, <strong>${escapeHtml(donor.full_name)}</strong>!</p>
      <p>Confirmamos a sua contribuição de <strong>${formatCurrency(amountCents)}</strong> para ${escapeHtml(settings.site_name)}.</p>
      <p>Que Deus abençoe você e sua família.</p>
    </div>
  `;

  await sendEmail({
    to: donor.email,
    subject: `Recibo de contribuição | ${settings.site_name}`,
    html,
  });
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}
