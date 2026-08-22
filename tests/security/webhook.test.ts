import { describe, it, expect, vi, beforeEach } from "vitest";
import { createHmac } from "node:crypto";

vi.mock("@/lib/supabase/admin", () => ({
  createAdminClient: vi.fn(),
}));

vi.mock("@/lib/stripe/server", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/stripe/server")>();
  return {
    ...actual,
    getStripe: vi.fn(),
  };
});

vi.mock("@/lib/email/resend", () => ({
  sendEmail: vi.fn().mockResolvedValue({ ok: true }),
}));

import { handleStripeWebhook } from "@/features/payments/webhook";
import { parseStripeSignature } from "@/lib/stripe/server";
import { createAdminClient } from "@/lib/supabase/admin";


function buildFakeSupabase(existingEventId?: string) {
  const inserted: Record<string, unknown>[] = [];

  const chain = (data: unknown) => ({
    select: () => chain(data),
    eq: () => chain(data),
    gte: () => chain(data),
    lte: () => chain(data),
    order: () => chain(data),
    limit: () => chain(data),
    single: () => Promise.resolve({ data, error: null }),
    maybeSingle: () =>
      Promise.resolve(
        data === null ? { data: null, error: null } : { data, error: null }
      ),
    insert: (values: unknown) => {
      if (typeof values === "object" && values !== null) {
        inserted.push(values as Record<string, unknown>);
      }
      return { error: null };
    },
    update: () => chain({}),
  });

  const from = (table: string) => {
    if (table === "stripe_webhook_events" && existingEventId) {
      return chain({ id: "existing" });
    }
    if (table === "recurring_donations") return chain({});
    if (table === "donors") return chain({ email: "a@b.com", full_name: "A" });
    return chain(null);
  };

  return { admin: { from } as never, inserted };
}

describe("Webhook Stripe", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("rejeita assinatura inválida", () => {
    const payload = JSON.stringify({ id: "evt_bad", type: "x" });
    const signature = `t=${Math.floor(Date.now() / 1000)},v1=invalida`;
    expect(parseStripeSignature(payload, signature)).toBeNull();
  });

  it("processa evento com assinatura válida", () => {
    const event = { id: "evt_valid", type: "payment_intent.succeeded", data: { object: {} } };
    const payload = JSON.stringify(event);
    const timestamp = Math.floor(Date.now() / 1000);
    const signedPayload = `${timestamp}.${payload}`;
    const signature = createHmac("sha256", process.env.STRIPE_WEBHOOK_SECRET!)
      .update(signedPayload)
      .digest("hex");
    const header = `t=${timestamp},v1=${signature}`;

    const result = parseStripeSignature(payload, header);
    expect(result).not.toBeNull();
    expect(result?.id).toBe("evt_valid");
  });

  it("é idempotente: ignora evento já processado", async () => {
    const { admin, inserted } = buildFakeSupabase("existing");
    vi.mocked(createAdminClient).mockReturnValue(admin);

    const result = await handleStripeWebhook({
      id: "evt_dup",
      type: "payment_intent.succeeded",
      data: { object: {} },
    } as never);

    expect(result.status).toBe("skipped");
    expect(inserted).toHaveLength(0);
  });

  it("registra evento novo com idempotência por evento", async () => {
    const { admin, inserted } = buildFakeSupabase();
    vi.mocked(createAdminClient).mockReturnValue(admin);

    await handleStripeWebhook({
      id: "evt_new",
      type: "payment_intent.succeeded",
      data: {
        object: {
          id: "pi_1",
          amount: 2000,
          metadata: { donor_id: "d1", community_id: "c1", payment_type: "one_time" },
        },
      },
    } as never);

    const webhookInsert = inserted.find((i) => "event_type" in i);
    expect(webhookInsert).toBeDefined();
    expect(webhookInsert!["stripe_event_id"]).toBe("evt_new");
  });
});