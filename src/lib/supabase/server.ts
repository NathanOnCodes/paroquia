import { createServerClient } from "@supabase/ssr";
import { cookies } from "next/headers";
import { env } from "@/lib/env";

export async function createClient() {
  const cookieStore = await cookies();

  return createServerClient(
    env.NEXT_PUBLIC_URL_SUPABASE,
    env.NEXT_PUBLIC_CHAVE_PUBLICA_SUPABASE,
    {
      cookies: {
        getAll() {
          return cookieStore.getAll();
        },
        setAll(cookiesToSet) {
          try {
            cookiesToSet.forEach(({ name, value, options }) =>
              cookieStore.set(name, value, options)
            );
          } catch {
            // setAll foi chamado a partir de um Server Component.
            // Isso pode ser ignorado se houver refresh de sessão via proxy.
          }
        },
      },
    }
  );
}
