import { createBrowserClient } from "@supabase/ssr";
import { env } from "@/lib/env";

export function createBrowserSupabaseClient() {
  return createBrowserClient(
    env.NEXT_PUBLIC_URL_SUPABASE,
    env.NEXT_PUBLIC_CHAVE_PUBLICA_SUPABASE
  );
}
