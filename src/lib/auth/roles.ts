import type { Role } from "@/lib/db-types";

/**
 * Rótulos exibidos na interface. Termos religiosos ficam somente aqui,
 * nunca como identificadores no banco/código.
 */
export const ROLE_LABELS: Record<Role, string> = {
  admin: "Admin",
  assistente: "Assistente",
};

export function hasRole(userRole: Role, required: Role): boolean {
  if (required === "admin") return userRole === "admin";
  if (required === "assistente") {
    return userRole === "admin" || userRole === "assistente";
  }
  return false;
}

export function canManageUsers(userRole: Role): boolean {
  return userRole === "admin";
}

export function canPublishTransparency(userRole: Role): boolean {
  return userRole === "admin";
}