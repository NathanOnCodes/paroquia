import { describe, it, expect } from "vitest";
import { titheFormSchema, createSubscriptionSchema } from "@/features/payments/schemas";
import { STRIPE_MIN_AMOUNT_CENTS } from "@/lib/stripe/server";

const base = {
  full_name: "João da Silva",
  email: "joao@example.com",
  phone: "(11) 99999-9999",
  community_id: "00000000-0000-0000-0000-000000000001",
  city: "São Paulo",
  neighborhood: "",
  amount_cents: 2000,
  is_recurring: false,
};

describe("Validação do formulário de dízimo", () => {
  it("aceita um valor válido com consentimento", () => {
    expect(() =>
      titheFormSchema.parse({ ...base, privacy_consent: true })
    ).not.toThrow();
  });

  it("rejeita sem consentimento de privacidade", () => {
    expect(() =>
      titheFormSchema.parse({ ...base, privacy_consent: false })
    ).toThrow();
  });

  it("rejeita valor abaixo do mínimo da Stripe", () => {
    expect(() =>
      titheFormSchema.parse({
        ...base,
        amount_cents: STRIPE_MIN_AMOUNT_CENTS - 1,
        privacy_consent: true,
      })
    ).toThrow();
  });

  it("rejeita valor não inteiro", () => {
    expect(() =>
      titheFormSchema.parse({
        ...base,
        amount_cents: 19.99,
        privacy_consent: true,
      })
    ).toThrow();
  });

  it("rejeita e-mail inválido", () => {
    expect(() =>
      titheFormSchema.parse({ ...base, email: "nao-e-email", privacy_consent: true })
    ).toThrow();
  });

  it("rejeita telefone inválido", () => {
    expect(() =>
      titheFormSchema.parse({ ...base, phone: "abc", privacy_consent: true })
    ).toThrow();
  });

  it("rejeita payload com recorrência fora do schema de pagamento avulso", () => {
    expect(() =>
      createSubscriptionSchema.parse({
        donor_id: "00000000-0000-0000-0000-000000000002",
        community_id: base.community_id,
        amount_cents: 2000,
        stripe_customer_id: "cus_123",
        payment_method_id: "",
      })
    ).toThrow();
  });
});