import { z } from "zod";

const envSchema = z.object({
  NEXT_PUBLIC_URL_SUPABASE: z.string().url().optional(),
  NEXT_PUBLIC_CHAVE_PUBLICA_SUPABASE: z.string().min(1).optional(),
  CHAVE_SERVICO_SUPABASE: z.string().min(1).optional(),
  NEXT_PUBLIC_CHAVE_PUBLICA_STRIPE: z.string().min(1).optional(),
  CHAVE_SECRETA_STRIPE: z.string().min(1).optional(),
  SEGREDO_WEBHOOK_STRIPE: z.string().min(1).optional(),
  CHAVE_API_RESEND: z.string().min(1).optional(),
  EMAIL_REMETENTE: z
    .string()
    .default("Paróquia <nao-responder@paroquia.local>"),
  EMAIL_NOTIFICACAO_ADMIN: z.string().email().optional(),
  NEXT_PUBLIC_URL_SITE: z.string().url().default("http://localhost:3000"),
  TTL_HORAS_TOKEN_RECORRENTE: z.coerce.number().min(1).default(24),
  TAMANHO_MAXIMO_ARQUIVO_MB: z.coerce.number().min(1).default(10),
});

function parseEnv() {
  const parsed = envSchema.safeParse(process.env);
  if (!parsed.success) {
    const missing = parsed.error.issues
      .map((i) => i.path.join("."))
      .join(", ");
    throw new Error(`Variáveis de ambiente inválidas: ${missing}`);
  }

  const required = [
    "NEXT_PUBLIC_URL_SUPABASE",
    "NEXT_PUBLIC_CHAVE_PUBLICA_SUPABASE",
    "CHAVE_SERVICO_SUPABASE",
    "NEXT_PUBLIC_CHAVE_PUBLICA_STRIPE",
    "CHAVE_SECRETA_STRIPE",
    "SEGREDO_WEBHOOK_STRIPE",
    "CHAVE_API_RESEND",
  ] as const;
  const missing = required.filter((key) => !parsed.data[key]);
  const isProduction = process.env.NODE_ENV === "production";

  if (isProduction && missing.length > 0) {
    throw new Error(`Variáveis de ambiente inválidas: ${missing.join(", ")}`);
  }

  return {
    NEXT_PUBLIC_URL_SUPABASE: parsed.data.NEXT_PUBLIC_URL_SUPABASE ?? "http://127.0.0.1:54321",
    NEXT_PUBLIC_CHAVE_PUBLICA_SUPABASE: parsed.data.NEXT_PUBLIC_CHAVE_PUBLICA_SUPABASE ?? "local-demo",
    CHAVE_SERVICO_SUPABASE: parsed.data.CHAVE_SERVICO_SUPABASE ?? "local-demo",
    NEXT_PUBLIC_CHAVE_PUBLICA_STRIPE: parsed.data.NEXT_PUBLIC_CHAVE_PUBLICA_STRIPE ?? "local-demo",
    CHAVE_SECRETA_STRIPE: parsed.data.CHAVE_SECRETA_STRIPE ?? "local-demo",
    SEGREDO_WEBHOOK_STRIPE: parsed.data.SEGREDO_WEBHOOK_STRIPE ?? "local-demo",
    CHAVE_API_RESEND: parsed.data.CHAVE_API_RESEND ?? "local-demo",
    EMAIL_REMETENTE: parsed.data.EMAIL_REMETENTE,
    EMAIL_NOTIFICACAO_ADMIN: parsed.data.EMAIL_NOTIFICACAO_ADMIN,
    NEXT_PUBLIC_URL_SITE: parsed.data.NEXT_PUBLIC_URL_SITE,
    TTL_HORAS_TOKEN_RECORRENTE: parsed.data.TTL_HORAS_TOKEN_RECORRENTE,
    TAMANHO_MAXIMO_ARQUIVO_MB: parsed.data.TAMANHO_MAXIMO_ARQUIVO_MB,
    LOCAL_DEMO_MODE: !isProduction && missing.length > 0,
  };
}

export const env = parseEnv();
