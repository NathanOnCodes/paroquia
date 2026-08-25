"use server";

import { revalidatePath } from "next/cache";
import { createAdminClient } from "@/lib/supabase/admin";
import { requireRole } from "@/lib/auth/authorization";
import { logAudit } from "@/lib/audit";
import { siteSettingsSchema } from "@/features/settings/schemas";
import { createStaticPix, hasError } from "pix-utils";

export async function updateSiteSettingsAction(formData: FormData): Promise<{
  error?: string;
  ok?: boolean;
}> {
  const actor = await requireRole("admin");
  const parsed = siteSettingsSchema.safeParse({
    site_name: formData.get("site_name"),
    site_title: formData.get("site_title"),
    city: formData.get("city") ?? "",
    description: formData.get("description"),
    secondary_color: formData.get("secondary_color"),
    pix_key: formData.get("pix_key") ?? "",
    pix_receiver_name: formData.get("pix_receiver_name") ?? "",
    pix_receiver_city: formData.get("pix_receiver_city") ?? "",
    home_title: formData.get("home_title") ?? "",
    home_description: formData.get("home_description") ?? "",
  });

  if (!parsed.success) {
    return { error: parsed.error.issues[0]?.message ?? "Dados inválidos" };
  }

  if (parsed.data.pix_key && (!parsed.data.pix_receiver_name || !parsed.data.pix_receiver_city)) {
    return { error: "Preencha nome e cidade do recebedor para configurar o Pix." };
  }

  if (parsed.data.pix_key) {
    const pix = createStaticPix({
      pixKey: parsed.data.pix_key,
      merchantName: parsed.data.pix_receiver_name,
      merchantCity: parsed.data.pix_receiver_city,
      transactionAmount: 0,
    });
    if (hasError(pix)) return { error: `Chave Pix inválida: ${pix.message}` };
  }

  const admin = createAdminClient();
  const { error } = await admin.from("site_settings").upsert({
    id: 1,
    ...parsed.data,
    pix_key: parsed.data.pix_key || null,
    pix_receiver_name: parsed.data.pix_receiver_name || null,
    pix_receiver_city: parsed.data.pix_receiver_city || null,
    home_title: parsed.data.home_title || null,
    home_description: parsed.data.home_description || null,
  });

  if (error) return { error: "Falha ao salvar as configurações" };

  await logAudit({
    actor_id: actor.id,
    action: "update_site_settings",
    entity_type: "site_settings",
    entity_id: "1",
  });
  revalidatePath("/", "layout");
  revalidatePath("/admin/configuracoes");
  return { ok: true };
}
