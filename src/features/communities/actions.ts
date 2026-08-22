"use server";

import { revalidatePath } from "next/cache";
import { createAdminClient } from "@/lib/supabase/admin";
import { communitySchema, communityStatusSchema } from "@/features/communities/schemas";
import { requireRole } from "@/lib/auth/authorization";
import { logAudit } from "@/lib/audit";

export async function createCommunityAction(formData: FormData): Promise<{
  error?: string;
  ok?: boolean;
}> {
  const actor = await requireRole("assistente");
  const parsed = communitySchema.safeParse({
    name: formData.get("name"),
    description: formData.get("description") ?? "",
    city: formData.get("city") ?? "",
    is_active: formData.get("is_active") === "on" || true,
  });

  if (!parsed.success) return { error: "Dados inválidos" };

  const admin = createAdminClient();
  const { data, error } = await admin
    .from("communities")
    .insert(parsed.data)
    .select("id")
    .single();
  if (error) return { error: "Falha ao criar comunidade" };

  await logAudit({
    actor_id: actor.id,
    action: "create_community",
    entity_type: "community",
    entity_id: data.id,
  });
  revalidatePath("/admin/comunidades");
  return { ok: true };
}

export async function updateCommunityAction(formData: FormData): Promise<{
  error?: string;
  ok?: boolean;
}> {
  const actor = await requireRole("assistente");
  const id = formData.get("id");
  if (typeof id !== "string" || !id) return { error: "Comunidade inválida" };

  const parsed = communitySchema.safeParse({
    name: formData.get("name"),
    description: formData.get("description") ?? "",
    city: formData.get("city") ?? "",
    is_active: formData.get("is_active") === "on" || true,
  });
  if (!parsed.success) return { error: "Dados inválidos" };

  const admin = createAdminClient();
  const { error } = await admin.from("communities").update(parsed.data).eq("id", id);
  if (error) return { error: "Falha ao atualizar comunidade" };

  await logAudit({
    actor_id: actor.id,
    action: "update_community",
    entity_type: "community",
    entity_id: id,
  });
  revalidatePath("/admin/comunidades");
  return { ok: true };
}

export async function setCommunityStatusAction(formData: FormData): Promise<{
  error?: string;
  ok?: boolean;
}> {
  const actor = await requireRole("assistente");
  const id = formData.get("id");
  const isActive = formData.get("is_active") === "on" || formData.get("is_active") === "true";

  if (typeof id !== "string" || !id) return { error: "Comunidade inválida" };
  const parsed = communityStatusSchema.safeParse({ is_active: isActive });
  if (!parsed.success) return { error: "Status inválido" };

  const admin = createAdminClient();
  const { error } = await admin
    .from("communities")
    .update({ is_active: parsed.data.is_active })
    .eq("id", id);
  if (error) return { error: "Falha ao atualizar status" };

  await logAudit({
    actor_id: actor.id,
    action: parsed.data.is_active ? "activate_community" : "deactivate_community",
    entity_type: "community",
    entity_id: id,
  });
  revalidatePath("/admin/comunidades");
  return { ok: true };
}