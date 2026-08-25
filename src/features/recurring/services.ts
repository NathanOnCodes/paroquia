import { createAdminClient } from "@/lib/supabase/admin";
import { hashValue, randomToken } from "@/lib/format";
import { sendEmail } from "@/lib/email/resend";
import { env } from "@/lib/env";
import { findDonorByEmail } from "@/features/donors/services";
import type { RecurringMagicToken } from "@/lib/db-types";

export async function createMagicToken(email: string): Promise<{
  rawToken: string;
}> {
  const rawToken = randomToken();
  const tokenHash = hashValue(rawToken);
  const ttlMs = env.TTL_HORAS_TOKEN_RECORRENTE * 60 * 60 * 1000;

  await createAdminClient().from("recurring_magic_tokens").insert({
    token_hash: tokenHash,
    email: email.toLowerCase(),
    expires_at: new Date(Date.now() + ttlMs).toISOString(),
  });

  return { rawToken };
}

export async function validateMagicToken(
  rawToken: string
): Promise<RecurringMagicToken | null> {
  const tokenHash = hashValue(rawToken);
  const { data } = await createAdminClient()
    .from("recurring_magic_tokens")
    .select("*")
    .eq("token_hash", tokenHash)
    .maybeSingle();

  if (!data) return null;
  const record = data as RecurringMagicToken;

  if (new Date(record.expires_at) < new Date()) return null;
  return record;
}

export async function consumeMagicToken(tokenId: string): Promise<void> {
  await createAdminClient()
    .from("recurring_magic_tokens")
    .update({ consumed_at: new Date().toISOString() })
    .eq("id", tokenId);
}

export async function sendRecurringManagementLink(email: string): Promise<void> {
  const donor = await findDonorByEmail(email);
  if (!donor) return;

  const { rawToken } = await createMagicToken(email);
  const link = `${env.NEXT_PUBLIC_URL_SITE}/api/recurring/session?token=${rawToken}`;

  const html = `
    <div style="font-family: Arial, sans-serif; max-width: 600px;">
      <h2>Gerenciar sua doação recorrente</h2>
      <p>Olá, <strong>${escapeHtml(donor.full_name)}</strong>!</p>
      <p>Você solicitou acesso para gerenciar suas doações recorrentes.</p>
      <p>Este link é válido por <strong>${env.TTL_HORAS_TOKEN_RECORRENTE} hora(s)</strong> e pode ser usado apenas uma vez.</p>
      <p><a href="${link}" style="background:#1976d2;color:#fff;padding:12px 20px;text-decoration:none;border-radius:6px;display:inline-block;">Acessar minhas recorrências</a></p>
      <p>Se não foi você quem solicitou, ignore este e-mail.</p>
    </div>
  `;

  await sendEmail({
    to: email,
    subject: "Gerenciar sua doação recorrente",
    html,
  });
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}
