"use server";

import { revalidatePath } from "next/cache";
import { createAdminClient } from "@/lib/supabase/admin";
import {
  financialEntrySchema,
  financialPeriodSchema,
  transparencyStatusSchema,
} from "@/features/transparency/schemas";
import { requireRole } from "@/lib/auth/authorization";
import { env } from "@/lib/env";
import { logAudit } from "@/lib/audit";

const DOCUMENT_BUCKET = "transparency";
const ALLOWED_MIME = new Set([
  "application/pdf",
  "image/jpeg",
  "image/png",
  "image/webp",
]);

export async function createFinancialPeriodAction(formData: FormData): Promise<{
  error?: string;
  ok?: boolean;
}> {
  const actor = await requireRole("assistente");
  const parsed = financialPeriodSchema.safeParse({
    name: formData.get("name"),
    start_date: formData.get("start_date"),
    end_date: formData.get("end_date"),
    status: formData.get("status") ?? "draft",
  });
  if (!parsed.success) return { error: "Dados inválidos. Verifique as datas e o nome." };

  const start = new Date(parsed.data.start_date);
  const end = new Date(parsed.data.end_date);
  if (start > end) return { error: "A data final deve ser posterior à inicial" };

  const admin = createAdminClient();
  const { data, error } = await admin
    .from("financial_periods")
    .insert({
      name: parsed.data.name,
      start_date: parsed.data.start_date,
      end_date: parsed.data.end_date,
      status: parsed.data.status,
      created_by: actor.id,
    })
    .select("id")
    .single();
  if (error) return { error: "Falha ao criar período" };

  await logAudit({
    actor_id: actor.id,
    action: "create_financial_period",
    entity_type: "financial_period",
    entity_id: data.id,
    metadata: { status: parsed.data.status },
  });
  revalidatePath("/admin/transparencia");
  return { ok: true };
}

export async function updateFinancialPeriodAction(formData: FormData): Promise<{
  error?: string;
  ok?: boolean;
}> {
  const actor = await requireRole("assistente");
  const id = formData.get("id");
  if (typeof id !== "string" || !id) return { error: "Período inválido" };

  const parsed = financialPeriodSchema.safeParse({
    name: formData.get("name"),
    start_date: formData.get("start_date"),
    end_date: formData.get("end_date"),
    status: formData.get("status") ?? "draft",
  });
  if (!parsed.success) return { error: "Dados inválidos" };

  const admin = createAdminClient();
  const { error } = await admin
    .from("financial_periods")
    .update({
      name: parsed.data.name,
      start_date: parsed.data.start_date,
      end_date: parsed.data.end_date,
    })
    .eq("id", id);
  if (error) return { error: "Falha ao atualizar período" };

  await logAudit({
    actor_id: actor.id,
    action: "update_financial_period",
    entity_type: "financial_period",
    entity_id: id,
  });
  revalidatePath("/admin/transparencia");
  return { ok: true };
}

export async function changeFinancialPeriodStatusAction(formData: FormData): Promise<{
  error?: string;
  ok?: boolean;
}> {
  const id = formData.get("id");
  const status = formData.get("status");
  if (typeof id !== "string" || !id) return { error: "Período inválido" };

  const parsed = transparencyStatusSchema.safeParse({ status });
  if (!parsed.success) return { error: "Status inválido" };

  const actor = await requireRole(
    parsed.data.status === "published" || parsed.data.status === "archived"
      ? "admin"
      : "assistente"
  );

  const admin = createAdminClient();
  const update: Record<string, unknown> = {
    status: parsed.data.status,
  };
  if (parsed.data.status === "published") {
    update.published_by = actor.id;
    update.published_at = new Date().toISOString();
  }
  if (parsed.data.status === "review") {
    update.reviewed_by = actor.id;
  }

  const { error } = await admin.from("financial_periods").update(update).eq("id", id);
  if (error) return { error: "Falha ao atualizar status. Apenas o pároco pode publicar." };

  await logAudit({
    actor_id: actor.id,
    action: "change_financial_period_status",
    entity_type: "financial_period",
    entity_id: id,
    metadata: { status: parsed.data.status },
  });
  revalidatePath("/admin/transparencia");
  revalidatePath("/transparencia");
  return { ok: true };
}

export async function addFinancialEntryAction(formData: FormData): Promise<{
  error?: string;
  ok?: boolean;
}> {
  const actor = await requireRole("assistente");
  const parsed = financialEntrySchema.safeParse({
    financial_period_id: formData.get("financial_period_id"),
    type: formData.get("type"),
    category: formData.get("category"),
    description: formData.get("description"),
    amount_cents: formData.get("amount_cents"),
    entry_date: formData.get("entry_date"),
  });
  if (!parsed.success) return { error: "Dados inválidos. Verifique os campos." };

  const admin = createAdminClient();
  const { error } = await admin.from("financial_entries").insert({
    ...parsed.data,
    source: "manual",
    created_by: actor.id,
  });
  if (error) return { error: "Falha ao adicionar lançamento" };

  await logAudit({
    actor_id: actor.id,
    action: "add_financial_entry",
    entity_type: "financial_entry",
    entity_id: parsed.data.financial_period_id,
    metadata: { type: parsed.data.type, amount_cents: parsed.data.amount_cents },
  });
  revalidatePath("/admin/transparencia");
  revalidatePath("/transparencia");
  return { ok: true };
}

export async function deleteFinancialEntryAction(formData: FormData): Promise<{
  error?: string;
  ok?: boolean;
}> {
  const actor = await requireRole("assistente");
  const id = formData.get("id");
  const periodId = formData.get("period_id");
  if (typeof id !== "string" || !id) return { error: "Lançamento inválido" };

  const admin = createAdminClient();
  const { error } = await admin.from("financial_entries").delete().eq("id", id);
  if (error) return { error: "Falha ao excluir lançamento" };

  await logAudit({
    actor_id: actor.id,
    action: "delete_financial_entry",
    entity_type: "financial_entry",
    entity_id: typeof periodId === "string" ? periodId : id,
  });
  revalidatePath("/admin/transparencia");
  revalidatePath("/transparencia");
  return { ok: true };
}

export async function uploadTransparencyDocumentAction(formData: FormData): Promise<{
  error?: string;
  ok?: boolean;
}> {
  const actor = await requireRole("assistente");
  const periodId = formData.get("period_id");
  if (typeof periodId !== "string" || !periodId) return { error: "Período inválido" };

  const file = formData.get("file");
  if (!(file instanceof File) || file.size === 0) {
    return { error: "Selecione um arquivo" };
  }
  if (!ALLOWED_MIME.has(file.type)) {
    return { error: "Formato inválido. Use PDF, JPG, PNG ou WebP." };
  }
  const maxBytes = env.MAX_FILE_SIZE_MB * 1024 * 1024;
  if (file.size > maxBytes) {
    return { error: `Arquivo acima do limite de ${env.MAX_FILE_SIZE_MB}MB` };
  }

  const bytes = new Uint8Array(await file.arrayBuffer());
  const ext = file.type === "application/pdf" ? "pdf" : file.type.split("/")[1];
  const path = `periods/${periodId}/${crypto.randomUUID()}.${ext}`;

  const admin = createAdminClient();
  const { error: uploadError } = await admin.storage
    .from(DOCUMENT_BUCKET)
    .upload(path, bytes, { contentType: file.type });
  if (uploadError) return { error: "Falha ao enviar arquivo" };

  const { error: dbError } = await admin.from("transparency_documents").insert({
    financial_period_id: periodId,
    file_path: path,
    original_file_name: file.name,
    mime_type: file.type,
    file_size: file.size,
    created_by: actor.id,
  });

  if (dbError) {
    await admin.storage.from(DOCUMENT_BUCKET).remove([path]);
    return { error: "Falha ao registrar documento" };
  }

  await logAudit({
    actor_id: actor.id,
    action: "upload_transparency_document",
    entity_type: "financial_period",
    entity_id: periodId,
  });
  revalidatePath("/admin/transparencia");
  revalidatePath("/transparencia");
  return { ok: true };
}