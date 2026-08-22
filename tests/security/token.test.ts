import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("@/lib/supabase/admin", () => ({
  createAdminClient: vi.fn(),
}));

vi.mock("@/lib/email/resend", () => ({
  sendEmail: vi.fn().mockResolvedValue({ ok: true }),
}));

import { validateMagicToken } from "@/features/recurring/services";
import { hashValue } from "@/lib/format";
import { createAdminClient } from "@/lib/supabase/admin";

function buildFakeAdmin(tokenRecord: unknown) {
  const from = (table: string) => {
    if (table === "recurring_magic_tokens") {
      return {
        select: () => ({
          eq: () => ({
            maybeSingle: () => Promise.resolve({ data: tokenRecord, error: null }),
          }),
        }),
        insert: () => ({ error: null }),
      };
    }
    if (table === "donors") {
      return {
        select: () => ({
          eq: () => ({
            maybeSingle: () => Promise.resolve({ data: null, error: null }),
          }),
        }),
      };
    }
    return { insert: () => ({ error: null }) };
  };
  return { from } as unknown as ReturnType<typeof createAdminClient>;
}

describe("Token de gerenciamento de recorrência", () => {
  beforeEach(() => vi.clearAllMocks());

  it("rejeita token expirado", async () => {
    vi.mocked(createAdminClient).mockReturnValue(
      buildFakeAdmin({
        id: "t1",
        token_hash: hashValue("token-valido"),
        email: "fiel@example.com",
        expires_at: new Date(Date.now() - 1000).toISOString(),
        consumed_at: null,
      })
    );
    const record = await validateMagicToken("token-valido");
    expect(record).toBeNull();
  });

  it("aceita token válido e não expirado", async () => {
    vi.mocked(createAdminClient).mockReturnValue(
      buildFakeAdmin({
        id: "t1",
        token_hash: hashValue("token-valido"),
        email: "fiel@example.com",
        expires_at: new Date(Date.now() + 60_000).toISOString(),
        consumed_at: null,
      })
    );
    const record = await validateMagicToken("token-valido");
    expect(record).not.toBeNull();
    expect(record?.email).toBe("fiel@example.com");
  });

  it("hash armazenado nunca é o token em texto puro", async () => {
    const raw = "segredo-super-secreto";
    expect(hashValue(raw)).not.toContain(raw);
  });
});