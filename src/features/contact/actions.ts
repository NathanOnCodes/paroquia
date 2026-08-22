"use server";

import { revalidatePath } from "next/cache";
import { createAdminClient } from "@/lib/supabase/admin";
import { contactStatusSchema } from "@/features/contact/schemas";
import { requireRole } from "@/lib/auth/authorization";
import { logAudit } from "@/lib/audit";

export async function setContactStatusAction(formData: FormData): Promise<{
  error?: string;
  ok?: boolean;
}> {
  const actor = await requireRole("assistente");
  const id = formData.get("id");
  const status = formData.get("status");

  if (typeof id !== "string" || !id) return { error: "Mensagem inválida" };
  const parsed = contactStatusSchema.safeParse({ status });
  if (!parsed.success) return { error: "Status inválido" };

  const admin = createAdminClient();
  const { error } = await admin
    .from("contact_messages")
    .update({ status: parsed.data.status })
    .eq("id", id);
  if (error) return { error: "Falha ao atualizar status" };

  await logAudit({
    actor_id: actor.id,
    action: "update_contact_status",
    entity_type: "contact_message",
    entity_id: id,
    metadata: { status: parsed.data.status },
  });
  revalidatePath("/admin/contatos");
  return { ok: true };
}