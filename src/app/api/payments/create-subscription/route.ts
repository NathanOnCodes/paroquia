import { NextResponse } from "next/server";
import { createSubscriptionSchema } from "@/features/payments/schemas";
import { createSubscription } from "@/features/payments/services";
import { jsonError, handleRouteError } from "@/lib/http";

export const runtime = "nodejs";

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const parsed = createSubscriptionSchema.safeParse(body);
    if (!parsed.success) {
      return jsonError("Dados inválidos", 400);
    }
    const result = await createSubscription(parsed.data);
    return NextResponse.json(result);
  } catch (error) {
    return handleRouteError(error);
  }
}