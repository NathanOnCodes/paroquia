import { NextResponse } from "next/server";
import { titheFormSchema } from "@/features/payments/schemas";
import { prepareDonor, createRecurringSetup } from "@/features/payments/services";
import { jsonError, handleRouteError } from "@/lib/http";

export const runtime = "nodejs";

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const parsed = titheFormSchema.safeParse(body);
    if (!parsed.success) {
      return jsonError("Dados inválidos", 400);
    }
    if (!parsed.data.is_recurring) {
      return jsonError("Use o endpoint de pagamento avulso", 400);
    }

    const donor = await prepareDonor(parsed.data);
    const result = await createRecurringSetup({
      ...parsed.data,
      donor_id: donor.id,
    });

    return NextResponse.json({ ...result, donor_id: donor.id });
  } catch (error) {
    return handleRouteError(error);
  }
}