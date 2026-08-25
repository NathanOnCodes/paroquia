import { createClient as createSupabaseClient } from "@supabase/supabase-js";
import { env } from "@/lib/env";

/**
 * Cliente privilegiado. Só deve ser usado em operações server-side
 * confiáveis (webhook Stripe, URLs assinadas, operações de manutenção).
 */
export function createAdminClient() {
  return createSupabaseClient(
    env.NEXT_PUBLIC_URL_SUPABASE,
    env.CHAVE_SERVICO_SUPABASE,
    {
      auth: {
        autoRefreshToken: false,
        persistSession: false,
      },
    }
  );
}
