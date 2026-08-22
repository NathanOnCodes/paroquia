import { NextResponse } from "next/server";
import { contactInputSchema } from "@/features/contact/schemas";
import { jsonError, handleRouteError } from "@/lib/http";
import { createAdminClient } from "@/lib/supabase/admin";
import { hashValue } from "@/lib/format";
import { sendEmail } from "@/lib/email/resend";
import { env } from "@/lib/env";

export const runtime = "nodejs";

const RATE_LIMIT_WINDOW_MS = 5 * 60 * 1000;
const RATE_LIMIT_MAX = 3;

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const parsed = contactInputSchema.safeParse(body);
    if (!parsed.success) {
      return jsonError("Dados inválidos", 400);
    }

    const ip = request.headers.get("x-forwarded-for")?.split(",")[0]?.trim() ?? "unknown";
    const ipHash = hashValue(ip);
    const supabase = createAdminClient();

    const since = new Date(Date.now() - RATE_LIMIT_WINDOW_MS).toISOString();
    const { count } = await supabase
      .from("contact_rate_limit")
      .select("id", { count: "exact", head: true })
      .eq("kind", "contact")
      .eq("ip_hash", ipHash)
      .gte("created_at", since);

    if ((count ?? 0) >= RATE_LIMIT_MAX) {
      return jsonError("Muitas solicitações. Tente novamente mais tarde.", 429);
    }

    await supabase.from("contact_rate_limit").insert({ kind: "contact", ip_hash: ipHash });

    const { error } = await supabase.from("contact_messages").insert({
      name: parsed.data.name,
      email: parsed.data.email,
      phone: parsed.data.phone || null,
      message: parsed.data.message,
      ip_hash: ipHash,
    });

    if (error) {
      return jsonError("Falha ao enviar mensagem", 500);
    }

    if (env.ADMIN_NOTIFICATION_EMAIL) {
      const html = `
        <div style="font-family: Arial, sans-serif;">
          <h3>Nova mensagem de contato</h3>
          <p><strong>Nome:</strong> ${escapeHtml(parsed.data.name)}</p>
          <p><strong>E-mail:</strong> ${escapeHtml(parsed.data.email)}</p>
          <p><strong>Telefone:</strong> ${escapeHtml(parsed.data.phone ?? "Não informado")}</p>
          <p><strong>Mensagem:</strong></p>
          <p>${escapeHtml(parsed.data.message)}</p>
        </div>
      `;
      await sendEmail({
        to: env.ADMIN_NOTIFICATION_EMAIL,
        subject: "Nova mensagem no site da Paróquia São Benedito e Menino Jesus",
        html,
      });
    }

    return NextResponse.json({ ok: true });
  } catch (error) {
    return handleRouteError(error);
  }
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}
