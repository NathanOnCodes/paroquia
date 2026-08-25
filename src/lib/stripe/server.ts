import Stripe from "stripe";
import { env } from "@/lib/env";

export const STRIPE_ACCOUNT_CURRENCY = "brl";
export const STRIPE_MIN_AMOUNT_CENTS = 100;

export function getStripe(): Stripe {
  return new Stripe(env.CHAVE_SECRETA_STRIPE, {
    apiVersion: "2026-07-29.dahlia",
    typescript: true,
  });
}

export function isValidDonationAmount(amountCents: number): boolean {
  return (
    Number.isInteger(amountCents) &&
    amountCents >= STRIPE_MIN_AMOUNT_CENTS &&
    amountCents <= 10_000_000
  );
}

export function parseStripeSignature(
  body: string,
  signature: string | null
): Stripe.Event | null {
  if (!signature) return null;
  try {
    const stripe = getStripe();
    return stripe.webhooks.constructEvent(
      body,
      signature,
    env.SEGREDO_WEBHOOK_STRIPE
    );
  } catch {
    return null;
  }
}
