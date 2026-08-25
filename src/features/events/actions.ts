"use server";

import { revalidatePath } from "next/cache";
import { createAdminClient } from "@/lib/supabase/admin";
import { eventInputSchema, eventStatusSchema } from "@/features/events/schemas";
import { requireRole } from "@/lib/auth/authorization";
import { slugify } from "@/lib/format";
import { env } from "@/lib/env";
import { logAudit } from "@/lib/audit";

const EVENT_IMAGE_BUCKET = "events";
const ALLOWED_MIME = new Set([
  "image/jpeg",
  "image/png",
  "image/webp",
]);

export async function createEventAction(formData: FormData): Promise<{
  error?: string;
  ok?: boolean;
}> {
  const actor = await requireRole("assistente");
  const parsed = eventInputSchema.safeParse({
    title: formData.get("title"),
    slug: formData.get("slug") || undefined,
    description: formData.get("description") ?? "",
    starts_at: formData.get("starts_at"),
    ends_at: formData.get("ends_at") || null,
    location: formData.get("location") ?? "",
    status: formData.get("status") ?? "draft",
  });
  if (!parsed.success) return { error: "Dados inválidos. Verifique os campos." };

  const file = formData.get("image");
  let imagePath: string | null = null;

  if (file instanceof File && file.size > 0) {
    if (!ALLOWED_MIME.has(file.type)) {
      return { error: "Formato de imagem inválido (use JPG, PNG ou WebP)" };
    }
    const maxBytes = env.TAMANHO_MAXIMO_ARQUIVO_MB * 1024 * 1024;
    if (file.size > maxBytes) {
      return { error: `Imagem acima do limite de ${env.TAMANHO_MAXIMO_ARQUIVO_MB}MB` };
    }
    imagePath = await uploadEventImage(file);
  }

  const admin = createAdminClient();
  const baseSlug = parsed.data.slug || slugify(parsed.data.title);
  const slug = await uniqueEventSlug(admin, baseSlug);

  const { data, error } = await admin
    .from("events")
    .insert({
      title: parsed.data.title,
      slug,
      description: parsed.data.description,
      image_path: imagePath,
      starts_at: new Date(parsed.data.starts_at).toISOString(),
      ends_at: parsed.data.ends_at ? new Date(parsed.data.ends_at).toISOString() : null,
      location: parsed.data.location,
      status: parsed.data.status,
      created_by: actor.id,
      published_at:
        parsed.data.status === "published" ? new Date().toISOString() : null,
    })
    .select("id")
    .single();

  if (error) {
    if (imagePath) await removeEventImage(imagePath);
    return { error: "Falha ao criar evento" };
  }

  await logAudit({
    actor_id: actor.id,
    action: "create_event",
    entity_type: "event",
    entity_id: data.id,
    metadata: { status: parsed.data.status },
  });
  revalidatePath("/admin/eventos");
  revalidatePath("/eventos");
  return { ok: true };
}

export async function updateEventAction(formData: FormData): Promise<{
  error?: string;
  ok?: boolean;
}> {
  const actor = await requireRole("assistente");
  const id = formData.get("id");
  if (typeof id !== "string" || !id) return { error: "Evento inválido" };

  const parsed = eventInputSchema.safeParse({
    title: formData.get("title"),
    slug: formData.get("slug") || undefined,
    description: formData.get("description") ?? "",
    starts_at: formData.get("starts_at"),
    ends_at: formData.get("ends_at") || null,
    location: formData.get("location") ?? "",
    status: formData.get("status") ?? "draft",
  });
  if (!parsed.success) return { error: "Dados inválidos" };

  const admin = createAdminClient();
  const file = formData.get("image");
  let imagePath: string | null | undefined;

  if (file instanceof File && file.size > 0) {
    if (!ALLOWED_MIME.has(file.type)) {
      return { error: "Formato de imagem inválido (use JPG, PNG ou WebP)" };
    }
    imagePath = await uploadEventImage(file);
  }

  const update: Record<string, unknown> = {
    title: parsed.data.title,
    description: parsed.data.description,
    starts_at: new Date(parsed.data.starts_at).toISOString(),
    ends_at: parsed.data.ends_at ? new Date(parsed.data.ends_at).toISOString() : null,
    location: parsed.data.location,
    status: parsed.data.status,
    updated_by: actor.id,
  };
  if (imagePath) update.image_path = imagePath;

  const { data: existing } = await admin
    .from("events")
    .select("slug, status, image_path")
    .eq("id", id)
    .single();

  const baseSlug = parsed.data.slug || slugify(parsed.data.title);
  if (existing && existing.slug !== baseSlug) {
    update.slug = await uniqueEventSlug(admin, baseSlug, id);
  }

  if (
    parsed.data.status === "published" &&
    existing?.status !== "published"
  ) {
    update.published_at = new Date().toISOString();
  }

  const { error } = await admin.from("events").update(update).eq("id", id);
  if (error) return { error: "Falha ao atualizar evento" };

  if (imagePath && existing?.image_path) {
    await removeEventImage(existing.image_path);
  }

  await logAudit({
    actor_id: actor.id,
    action: "update_event",
    entity_type: "event",
    entity_id: id,
    metadata: { status: parsed.data.status },
  });
  revalidatePath("/admin/eventos");
  revalidatePath("/eventos");
  return { ok: true };
}

export async function setEventStatusAction(formData: FormData): Promise<{
  error?: string;
  ok?: boolean;
}> {
  const actor = await requireRole("assistente");
  const id = formData.get("id");
  const status = formData.get("status");

  if (typeof id !== "string" || !id) return { error: "Evento inválido" };
  const parsed = eventStatusSchema.safeParse({ status });
  if (!parsed.success) return { error: "Status inválido" };

  const admin = createAdminClient();
  const update: Record<string, unknown> = { status: parsed.data.status };
  if (parsed.data.status === "published") update.published_at = new Date().toISOString();
  if (parsed.data.status === "draft" || parsed.data.status === "archived") {
    update.published_at = null;
  }

  const { error } = await admin.from("events").update(update).eq("id", id);
  if (error) return { error: "Falha ao atualizar status" };

  await logAudit({
    actor_id: actor.id,
    action: "update_event_status",
    entity_type: "event",
    entity_id: id,
    metadata: { status: parsed.data.status },
  });
  revalidatePath("/admin/eventos");
  revalidatePath("/eventos");
  return { ok: true };
}

async function uploadEventImage(file: File): Promise<string> {
  const bytes = new Uint8Array(await file.arrayBuffer());
  const ext = file.type.split("/")[1] ?? "jpg";
  const path = `events/${crypto.randomUUID()}.${ext}`;

  const admin = createAdminClient();
  const { error } = await admin.storage
    .from(EVENT_IMAGE_BUCKET)
    .upload(path, bytes, { contentType: file.type });

  if (error) throw new Error("Falha ao enviar imagem");
  return path;
}

async function removeEventImage(path: string): Promise<void> {
  const admin = createAdminClient();
  await admin.storage.from(EVENT_IMAGE_BUCKET).remove([path]);
}

async function uniqueEventSlug(
  admin: ReturnType<typeof createAdminClient>,
  base: string,
  excludeId?: string
): Promise<string> {
  let slug = base;
  let counter = 1;
  while (true) {
    let query = admin.from("events").select("id").eq("slug", slug);
    if (excludeId) query = query.neq("id", excludeId);
    const { data } = await query.maybeSingle();
    if (!data) return slug;
    counter += 1;
    slug = `${base}-${counter}`;
  }
}
