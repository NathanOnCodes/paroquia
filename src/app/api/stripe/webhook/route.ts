import { NextResponse } from "next/server";
import { parseStripeSignature } from "@/lib/stripe/server";
import { handleStripeWebhook } from "@/features/payments/webhook";

export const runtime = "nodejs";

export async function POST(request: Request) {
  const body = await request.text();
  const signature = request.headers.get("stripe-signature");

  const event = parseStripeSignature(body, signature);
  if (!event) {
    return NextResponse.json({ error: "Assinatura inválida" }, { status: 400 });
  }

  try {
    const result = await handleStripeWebhook(event);
    return NextResponse.json({ received: true, ...result });
  } catch {
    return NextResponse.json({ error: "Falha ao processar evento" }, { status: 500 });
  }
}