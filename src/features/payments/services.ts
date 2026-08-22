import Stripe from "stripe";
import { getStripe, isValidDonationAmount } from "@/lib/stripe/server";
import { createAdminClient } from "@/lib/supabase/admin";
import { upsertDonor } from "@/features/donors/services";
import type { TitheFormInput } from "@/features/payments/schemas";
import type {
  CreateSubscriptionInput,
} from "@/features/payments/schemas";
import type {
  ContributionPayment,
  Donor,
  RecurringDonation,
} from "@/lib/db-types";

const PAYMENT_METHOD_TYPES = ["card", "pix"] as const;

type StripeTitheInput = TitheFormInput;

export async function prepareDonor(
  input: StripeTitheInput
): Promise<Donor> {
  return upsertDonor({
    full_name: input.full_name,
    email: input.email,
    phone: input.phone,
    city: input.city,
    neighborhood: input.neighborhood,
    community_id: input.community_id ?? null,
    community_name: input.community_id ? "" : input.community_name,
  });
}

function metadataFor(input: StripeTitheInput, donorId: string) {
  return {
    donor_id: donorId,
    community_id: input.community_id ?? "",
    community_name: input.community_id ? "" : input.community_name ?? "",
    amount_cents: String(input.amount_cents),
    payment_type: input.is_recurring ? "recurring" : "one_time",
  };
}

/**
 * Pagamento avulso: cria um PaymentIntent com métodos automáticos
 * (cartão, débito e PIX conforme habilitados na conta).
 */
export async function createOneTimePayment(
  input: StripeTitheInput & { donor_id: string }
): Promise<{ client_secret: string; amount_cents: number }> {
  if (!isValidDonationAmount(input.amount_cents)) {
    throw new Error("Valor inválido para pagamento");
  }
  const stripe = getStripe();
  const paymentIntent = await stripe.paymentIntents.create({
    amount: input.amount_cents,
    currency: "brl",
    automatic_payment_methods: { enabled: true },
    metadata: metadataFor(input, input.donor_id),
    receipt_email: input.email,
  });
  return { client_secret: paymentIntent.client_secret ?? "", amount_cents: input.amount_cents };
}

/**
 * Recorrência: garante o customer na Stripe e cria um SetupIntent
 * para coletar o mandato (cartão ou Pix Automático).
 */
export async function createRecurringSetup(
  input: StripeTitheInput & { donor_id: string }
): Promise<{
  client_secret: string;
  stripe_customer_id: string;
}> {
  const stripe = getStripe();

  const { data: existing } = await createAdminClient()
    .from("recurring_donations")
    .select("stripe_customer_id")
    .eq("donor_id", input.donor_id)
    .order("created_at", { ascending: false })
    .limit(1)
    .maybeSingle();

  let customerId = existing?.stripe_customer_id ?? null;

  if (!customerId) {
    const customer = await stripe.customers.create({
      email: input.email,
      name: input.full_name,
      phone: input.phone ?? undefined,
      metadata: { donor_id: input.donor_id },
    });
    customerId = customer.id;
  }

  const setupIntent = await stripe.setupIntents.create({
    customer: customerId,
    payment_method_types: [...PAYMENT_METHOD_TYPES],
    usage: "off_session",
    metadata: metadataFor(input, input.donor_id),
  });

  return {
    client_secret: setupIntent.client_secret ?? "",
    stripe_customer_id: customerId,
  };
}

/**
 * Cria a assinatura recorrente após o mandato ser coletado no SetupIntent.
 * A primeira cobrança é feita imediatamente via invoice.
 */
export async function createSubscription(
  input: CreateSubscriptionInput
): Promise<{ subscription_id: string }> {
  const stripe = getStripe();

  const price = await stripe.prices.create({
    currency: "brl",
    unit_amount: input.amount_cents,
    recurring: { interval: "month" },
    product_data: { name: "Dízimo recorrente mensal" },
    metadata: { donor_id: input.donor_id },
  });

  const subscription = await stripe.subscriptions.create({
    customer: input.stripe_customer_id,
    items: [{ price: price.id }],
    payment_behavior: "default_incomplete",
    default_payment_method: input.payment_method_id,
    payment_settings: {
      payment_method_types: [...PAYMENT_METHOD_TYPES],
      save_default_payment_method: "on_subscription",
    },
    metadata: {
      donor_id: input.donor_id,
      community_id: input.community_id ?? "",
      community_name: input.community_id ? "" : input.community_name ?? "",
      amount_cents: String(input.amount_cents),
      payment_type: "recurring",
    },
  });

  const admin = createAdminClient();
  const firstItem = subscription.items?.data?.[0];
  await admin.from("recurring_donations").insert({
    donor_id: input.donor_id,
    community_id: input.community_id ?? null,
    amount_cents: input.amount_cents,
    currency: "brl",
    interval: "month",
    stripe_customer_id: input.stripe_customer_id,
    stripe_subscription_id: subscription.id,
    status: "incomplete",
    started_at: firstItem?.current_period_start
      ? new Date(firstItem.current_period_start * 1000).toISOString()
      : null,
    current_period_end: firstItem?.current_period_end
      ? new Date(firstItem.current_period_end * 1000).toISOString()
      : null,
  });

  return { subscription_id: subscription.id };
}

/**
 * Cancela a recorrência ao final do ciclo atual (cancel_at_period_end).
 */
export async function cancelRecurringAtPeriodEnd(
  donorId: string,
  subscriptionId: string
): Promise<RecurringDonation> {
  const admin = createAdminClient();
  const { data: record } = await admin
    .from("recurring_donations")
    .select("*")
    .eq("stripe_subscription_id", subscriptionId)
    .eq("donor_id", donorId)
    .maybeSingle();

  if (!record) {
    throw new Error("Recorrência não encontrada");
  }
  if (record.status === "canceled") {
    return record as RecurringDonation;
  }

  const stripe = getStripe();
  await stripe.subscriptions.update(subscriptionId, {
    cancel_at_period_end: true,
  });

  const { data, error } = await admin
    .from("recurring_donations")
    .update({
      cancel_at_period_end: true,
      status: "cancel_at_period_end",
      cancellation_requested_at: new Date().toISOString(),
    })
    .eq("id", record.id)
    .select("*")
    .single();

  if (error) throw new Error("Falha ao atualizar recorrência");
  return data as RecurringDonation;
}

export async function listRecurringForDonor(
  donorId: string
): Promise<RecurringDonation[]> {
  const admin = createAdminClient();
  const { data } = await admin
    .from("recurring_donations")
    .select("*")
    .eq("donor_id", donorId)
    .order("created_at", { ascending: false });
  return (data as RecurringDonation[]) ?? [];
}

export async function findContributionByPaymentIntent(
  paymentIntentId: string
): Promise<ContributionPayment | null> {
  const admin = createAdminClient();
  const { data } = await admin
    .from("contribution_payments")
    .select("*")
    .eq("stripe_payment_intent_id", paymentIntentId)
    .maybeSingle();
  return (data as ContributionPayment | null) ?? null;
}

export async function findContributionByInvoice(
  invoiceId: string
): Promise<ContributionPayment | null> {
  const admin = createAdminClient();
  const { data } = await admin
    .from("contribution_payments")
    .select("*")
    .eq("stripe_invoice_id", invoiceId)
    .maybeSingle();
  return (data as ContributionPayment | null) ?? null;
}

export async function findRecurringBySubscription(
  subscriptionId: string
): Promise<RecurringDonation | null> {
  const admin = createAdminClient();
  const { data } = await admin
    .from("recurring_donations")
    .select("*")
    .eq("stripe_subscription_id", subscriptionId)
    .maybeSingle();
  return (data as RecurringDonation | null) ?? null;
}

export type { Stripe };

export { PAYMENT_METHOD_TYPES };