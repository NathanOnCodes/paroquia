import { NextResponse } from "next/server";
import {
  createPaymentIntentSchema,
  titheFormSchema,
} from "@/features/payments/schemas";
import { prepareDonor, createOneTimePayment } from "@/features/payments/services";
import { jsonError, handleRouteError } from "@/lib/http";

export const runtime = "nodejs";

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const parsed = titheFormSchema.safeParse(body);
    if (!parsed.success) {
      return jsonError("Dados inválidos", 400);
    }
    if (parsed.data.is_recurring) {
      return jsonError("Use o endpoint de recorrência", 400);
    }
    createPaymentIntentSchema.parse(parsed.data);

    const donor = await prepareDonor(parsed.data);
    const result = await createOneTimePayment({
      ...parsed.data,
      donor_id: donor.id,
    });

    return NextResponse.json(result);
  } catch (error) {
    return handleRouteError(error);
  }
}