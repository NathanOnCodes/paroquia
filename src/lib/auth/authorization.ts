import { createClient } from "@/lib/supabase/server";
import { hasRole } from "@/lib/auth/roles";
import type { Profile, Role } from "@/lib/db-types";

export type AuthUser = {
  id: string;
  email: string;
  profile: Profile | null;
};

export type AuthedUser = Omit<AuthUser, "profile"> & { profile: Profile };

export async function getSessionUser(): Promise<AuthUser | null> {
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();

  if (!user) return null;

  const { data: profile } = await supabase
    .from("profiles")
    .select("*")
    .eq("id", user.id)
    .single();

  return { id: user.id, email: user.email ?? "", profile };
}

/**
 * Exige usuário autenticado e ativo. Lança erro se não atender.
 */
export async function requireUser(): Promise<AuthedUser> {
  const user = await getSessionUser();
  if (!user) {
    throw new Error("Não autenticado");
  }
  if (!user.profile) {
    throw new Error("Perfil não encontrado");
  }
  if (!user.profile.is_active) {
    throw new Error("Usuário inativo");
  }
  return user as AuthedUser;
}

/**
 * Exige papel mínimo. `assistente` aceita admin e assistentes;
 * `admin` aceita somente o admin (pároco na interface).
 */
export async function requireRole(required: Role): Promise<AuthedUser> {
  const user = await requireUser();
  if (!hasRole(user.profile.role, required)) {
    throw new Error("Sem permissão para esta operação");
  }
  return user;
}

/**
 * Verificação leve para uso em Server Components/páginas.
 * Retorna null em vez de lançar.
 */
export async function getAdminUser(required?: Role): Promise<AuthUser | null> {
  try {
    return required ? await requireRole(required) : await requireUser();
  } catch {
    return null;
  }
}