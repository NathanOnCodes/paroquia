"use server";

import { redirect } from "next/navigation";
import { createClient } from "@/lib/supabase/server";
import { createAdminClient } from "@/lib/supabase/admin";
import {
  createStaffUserSchema,
  loginSchema,
  updateStaffUserSchema,
} from "@/features/auth/schemas";
import { requireRole } from "@/lib/auth/authorization";
import { logAudit } from "@/lib/audit";

export async function loginAction(formData: FormData): Promise<{ error?: string }> {
  const parsed = loginSchema.safeParse({
    email: formData.get("email"),
    password: formData.get("password"),
  });

  if (!parsed.success) {
    return { error: "E-mail ou senha inválidos" };
  }

  const supabase = await createClient();
  const { error } = await supabase.auth.signInWithPassword({
    email: parsed.data.email,
    password: parsed.data.password,
  });

  if (error) {
    return { error: "E-mail ou senha inválidos" };
  }

  redirect("/admin");
}

export async function logoutAction(): Promise<void> {
  const supabase = await createClient();
  await supabase.auth.signOut();
  redirect("/");
}

export async function createStaffUserAction(formData: FormData): Promise<{
  error?: string;
  ok?: boolean;
}> {
  const actor = await requireRole("admin");
  const parsed = createStaffUserSchema.safeParse({
    email: formData.get("email"),
    full_name: formData.get("full_name"),
    password: formData.get("password"),
    role: formData.get("role"),
    is_active: formData.get("is_active") === "on" || true,
  });

  if (!parsed.success) {
    return { error: "Dados inválidos. Verifique os campos." };
  }

  const admin = createAdminClient();
  const { data: authUser, error: authError } = await admin.auth.admin.createUser({
    email: parsed.data.email,
    password: parsed.data.password,
    email_confirm: true,
  });

  if (authError) {
    return { error: "Falha ao criar usuário: " + authError.message };
  }

  const { error: profileError } = await admin.from("profiles").insert({
    id: authUser.user.id,
    full_name: parsed.data.full_name,
    email: parsed.data.email.toLowerCase(),
    role: parsed.data.role,
    is_active: parsed.data.is_active,
  });

  if (profileError) {
    await admin.auth.admin.deleteUser(authUser.user.id);
    return { error: "Falha ao criar perfil do usuário" };
  }

  await logAudit({
    actor_id: actor.id,
    action: "create_staff_user",
    entity_type: "profile",
    entity_id: authUser.user.id,
    metadata: { role: parsed.data.role },
  });

  return { ok: true };
}

export async function updateStaffUserAction(formData: FormData): Promise<{
  error?: string;
  ok?: boolean;
}> {
  const actor = await requireRole("admin");
  const id = formData.get("id");
  if (typeof id !== "string" || !id) return { error: "Usuário inválido" };

  const parsed = updateStaffUserSchema.safeParse({
    full_name: formData.get("full_name"),
    role: formData.get("role"),
    is_active: formData.get("is_active") === "on",
  });

  if (!parsed.success) return { error: "Dados inválidos" };
  if (id === actor.id && !parsed.data.is_active) {
    return { error: "Você não pode desativar o próprio usuário" };
  }

  const admin = createAdminClient();
  const { error } = await admin
    .from("profiles")
    .update({
      full_name: parsed.data.full_name,
      role: parsed.data.role,
      is_active: parsed.data.is_active,
    })
    .eq("id", id);

  if (error) return { error: "Falha ao atualizar usuário" };

  await logAudit({
    actor_id: actor.id,
    action: "update_staff_user",
    entity_type: "profile",
    entity_id: id,
    metadata: { role: parsed.data.role, is_active: parsed.data.is_active },
  });

  return { ok: true };
}