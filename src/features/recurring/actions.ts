"use server";

import { cookies } from "next/headers";
import { revalidatePath } from "next/cache";
import { validateMagicToken } from "@/features/recurring/services";
import {
  cancelRecurringAtPeriodEnd,
  listRecurringForDonor,
} from "@/features/payments/services";
import { findDonorByEmail } from "@/features/donors/services";
import { createAdminClient } from "@/lib/supabase/admin";
import { logAudit } from "@/lib/audit";

export async function getRecurringSession(): Promise<{
  donorId: string | null;
  subscriptions: Awaited<ReturnType<typeof listRecurringForDonor>>;
}> {
  const cookieStore = await cookies();
  const token = cookieStore.get("recurring_mgmt")?.value;
  if (!token) {
    return { donorId: null, subscriptions: [] };
  }

  const record = await validateMagicToken(token);
  if (!record) {
    return { donorId: null, subscriptions: [] };
  }

  const donor = await findDonorByEmail(record.email);
  if (!donor) {
    return { donorId: null, subscriptions: [] };
  }

  const subscriptions = await listRecurringForDonor(donor.id);
  return { donorId: donor.id, subscriptions };
}

export async function cancelRecurringAction(formData: FormData): Promise<{
  error?: string;
  ok?: boolean;
}> {
  const subscriptionId = formData.get("subscription_id");
  if (typeof subscriptionId !== "string" || !subscriptionId) {
    return { error: "Recorrência inválida" };
  }

  const cookieStore = await cookies();
  const token = cookieStore.get("recurring_mgmt")?.value;
  if (!token) {
    return { error: "Sessão expirada. Solicite um novo link de acesso." };
  }

  const record = await validateMagicToken(token);
  if (!record) {
    return { error: "Sessão expirada. Solicite um novo link de acesso." };
  }

  const donor = await findDonorByEmail(record.email);
  if (!donor) {
    return { error: "Dizimista não encontrado" };
  }

  const admin = createAdminClient();
  const { data: sub } = await admin
    .from("recurring_donations")
    .select("id, donor_id, status")
    .eq("stripe_subscription_id", subscriptionId)
    .eq("donor_id", donor.id)
    .maybeSingle();

  if (!sub) {
    return { error: "Recorrência não encontrada para este e-mail" };
  }

  try {
    await cancelRecurringAtPeriodEnd(donor.id, subscriptionId);
  } catch {
    return { error: "Falha ao cancelar a recorrência. Tente novamente." };
  }

  await logAudit({
    actor_id: donor.id,
    action: "donor_cancel_recurring",
    entity_type: "recurring_donation",
    entity_id: sub.id,
    metadata: { source: "public" },
  });

  revalidatePath("/doacoes/gerenciar");
  return { ok: true };
}
