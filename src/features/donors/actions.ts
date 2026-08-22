"use server";

import { revalidatePath } from "next/cache";
import { createAdminClient } from "@/lib/supabase/admin";
import { donorInputSchema, donorStatusSchema } from "@/features/donors/schemas";
import { requireRole } from "@/lib/auth/authorization";
import { logAudit } from "@/lib/audit";

export async function createDonorAction(formData: FormData): Promise<{
  error?: string;
  ok?: boolean;
}> {
  const actor = await requireRole("assistente");
  const parsed = donorInputSchema.safeParse({
    full_name: formData.get("full_name"),
    email: formData.get("email"),
    phone: formData.get("phone") ?? "",
    city: formData.get("city") ?? "",
    neighborhood: formData.get("neighborhood") ?? "",
    community_id: formData.get("community_id") || null,
    community_name: formData.get("community_name") ?? "",
  });

  if (!parsed.success) return { error: "Dados inválidos. Verifique os campos." };

  const admin = createAdminClient();
  const { data, error } = await admin
    .from("donors")
    .insert({
      ...parsed.data,
      phone: parsed.data.phone || null,
      community_id: parsed.data.community_id ?? null,
      community_name: parsed.data.community_id
        ? null
        : parsed.data.community_name || null,
      privacy_consent_at: new Date().toISOString(),
    })
    .select("id")
    .single();
  if (error) return { error: "Falha ao cadastrar dizimista" };

  await logAudit({
    actor_id: actor.id,
    action: "create_donor",
    entity_type: "donor",
    entity_id: data.id,
  });
  revalidatePath("/admin/dizimistas");
  return { ok: true };
}

export async function updateDonorAction(formData: FormData): Promise<{
  error?: string;
  ok?: boolean;
}> {
  const actor = await requireRole("assistente");
  const id = formData.get("id");
  if (typeof id !== "string" || !id) return { error: "Dizimista inválido" };

  const parsed = donorInputSchema.safeParse({
    full_name: formData.get("full_name"),
    email: formData.get("email"),
    phone: formData.get("phone") ?? "",
    city: formData.get("city") ?? "",
    neighborhood: formData.get("neighborhood") ?? "",
    community_id: formData.get("community_id") || null,
    community_name: formData.get("community_name") ?? "",
  });
  if (!parsed.success) return { error: "Dados inválidos" };

  const admin = createAdminClient();
  const { error } = await admin
    .from("donors")
    .update({
      ...parsed.data,
      phone: parsed.data.phone || null,
      community_id: parsed.data.community_id ?? null,
      community_name: parsed.data.community_id
        ? null
        : parsed.data.community_name || null,
    })
    .eq("id", id);
  if (error) return { error: "Falha ao atualizar dizimista" };

  await logAudit({
    actor_id: actor.id,
    action: "update_donor",
    entity_type: "donor",
    entity_id: id,
  });
  revalidatePath("/admin/dizimistas");
  return { ok: true };
}

export async function setDonorStatusAction(formData: FormData): Promise<{
  error?: string;
  ok?: boolean;
}> {
  const actor = await requireRole("assistente");
  const id = formData.get("id");
  const isActive = formData.get("is_active") === "on" || formData.get("is_active") === "true";

  if (typeof id !== "string" || !id) return { error: "Dizimista inválido" };
  const parsed = donorStatusSchema.safeParse({ is_active: isActive });
  if (!parsed.success) return { error: "Status inválido" };

  const admin = createAdminClient();
  const { error } = await admin
    .from("donors")
    .update({ is_active: parsed.data.is_active })
    .eq("id", id);
  if (error) return { error: "Falha ao atualizar status" };

  await logAudit({
    actor_id: actor.id,
    action: parsed.data.is_active ? "activate_donor" : "deactivate_donor",
    entity_type: "donor",
    entity_id: id,
  });
  revalidatePath("/admin/dizimistas");
  return { ok: true };
}