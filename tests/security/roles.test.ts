import { describe, it, expect } from "vitest";
import { hasRole } from "@/lib/auth/roles";
import { canManageUsers, canPublishTransparency } from "@/lib/auth/roles";

describe("Autorização por papel", () => {
  it("assistente não possui acesso exclusivo de admin", () => {
    expect(hasRole("assistente", "admin")).toBe(false);
  });

  it("admin possui acesso de assistente e de admin", () => {
    expect(hasRole("admin", "admin")).toBe(true);
    expect(hasRole("admin", "assistente")).toBe(true);
  });

  it("assistente possui acesso de assistente", () => {
    expect(hasRole("assistente", "assistente")).toBe(true);
  });

  it("apenas admin gerencia usuários", () => {
    expect(canManageUsers("admin")).toBe(true);
    expect(canManageUsers("assistente")).toBe(false);
  });

  it("apenas admin publica transparência", () => {
    expect(canPublishTransparency("admin")).toBe(true);
    expect(canPublishTransparency("assistente")).toBe(false);
  });
});