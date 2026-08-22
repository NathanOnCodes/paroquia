import { z } from "zod";

const envSchema = z.object({
  NEXT_PUBLIC_SUPABASE_URL: z.string().url(),
  NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY: z.string().min(1),
  SUPABASE_SERVICE_ROLE_KEY: z.string().min(1),
  NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY: z.string().min(1),
  STRIPE_SECRET_KEY: z.string().min(1),
  STRIPE_WEBHOOK_SECRET: z.string().min(1),
  RESEND_API_KEY: z.string().min(1),
  EMAIL_FROM: z
    .string()
    .default("Paróquia São Benedito e Menino Jesus <nao-responder@paroquia.local>"),
  ADMIN_NOTIFICATION_EMAIL: z.string().email().optional(),
  NEXT_PUBLIC_SITE_URL: z.string().url().default("http://localhost:3000"),
  RECURRING_TOKEN_TTL_HOURS: z.coerce.number().min(1).default(24),
  MAX_FILE_SIZE_MB: z.coerce.number().min(1).default(10),
});

function parseEnv() {
  const parsed = envSchema.safeParse(process.env);
  if (!parsed.success) {
    const missing = parsed.error.issues
      .map((i) => i.path.join("."))
      .join(", ");
    throw new Error(`Variáveis de ambiente inválidas: ${missing}`);
  }
  return parsed.data;
}

export const env = parseEnv();
