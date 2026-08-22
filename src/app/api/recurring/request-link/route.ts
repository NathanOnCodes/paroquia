import { NextResponse } from "next/server";
import { requestRecurringLinkSchema } from "@/features/payments/schemas";
import { jsonError, handleRouteError } from "@/lib/http";
import { createAdminClient } from "@/lib/supabase/admin";
import { hashValue } from "@/lib/format";
import { sendRecurringManagementLink } from "@/features/recurring/services";

export const runtime = "nodejs";

const WINDOW_MS = 5 * 60 * 1000;
const MAX_PER_IP = 3;
const MAX_PER_EMAIL = 2;

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const parsed = requestRecurringLinkSchema.safeParse(body);
    if (!parsed.success) {
      return jsonError("Dados inválidos", 400);
    }

    const ip =
      request.headers.get("x-forwarded-for")?.split(",")[0]?.trim() ?? "unknown";
    const ipHash = hashValue(ip);
    const email = parsed.data.email;
    const since = new Date(Date.now() - WINDOW_MS).toISOString();
    const supabase = createAdminClient();

    const { count: ipCount } = await supabase
      .from("contact_rate_limit")
      .select("id", { count: "exact", head: true })
      .eq("kind", "recurring_link")
      .eq("ip_hash", ipHash)
      .gte("created_at", since);

    if ((ipCount ?? 0) >= MAX_PER_IP) {
      return jsonError("Muitas solicitações. Tente novamente mais tarde.", 429);
    }

    const { count: emailCount } = await supabase
      .from("recurring_magic_tokens")
      .select("id", { count: "exact", head: true })
      .eq("email", email)
      .gte("created_at", since);

    if ((emailCount ?? 0) >= MAX_PER_EMAIL) {
      return jsonError("Muitas solicitações. Tente novamente mais tarde.", 429);
    }

    await supabase.from("contact_rate_limit").insert({
      kind: "recurring_link",
      ip_hash: ipHash,
    });

    // Resposta genérica independente de o e-mail existir.
    await sendRecurringManagementLink(email);

    return NextResponse.json({
      ok: true,
      message: "Se o e-mail estiver cadastrado, você receberá um link de acesso.",
    });
  } catch (error) {
    return handleRouteError(error);
  }
}