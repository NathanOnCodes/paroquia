import { z } from "zod";
import { STRIPE_MIN_AMOUNT_CENTS } from "@/lib/stripe/server";

function requireCommunity(
  data: { community_id?: string | null; community_name?: string },
  ctx: z.RefinementCtx
) {
  if (!data.community_id && !data.community_name) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      path: ["community_id"],
      message: "Informe a sua comunidade",
    });
  }
}

const titheFormBase = z.object({
  full_name: z.string().trim().min(3, "Informe o nome completo").max(200),
  email: z.string().trim().toLowerCase().email("E-mail inválido").max(200),
  phone: z
    .string()
    .trim()
    .regex(/^[0-9()\s-]{8,20}$/, "Telefone inválido"),
  community_id: z.string().uuid("Comunidade inválida").nullable().optional(),
  community_name: z
    .string()
    .trim()
    .max(160, "Nome muito longo")
    .optional()
    .default(""),
  city: z.string().trim().min(2, "Informe a cidade").max(120),
  neighborhood: z.string().trim().max(120).optional().default(""),
  amount_cents: z.coerce
    .number()
    .int("Valor inválido")
    .min(STRIPE_MIN_AMOUNT_CENTS, "Valor mínimo de R$ 1,00")
    .max(10_000_000, "Valor acima do permitido"),
  is_recurring: z.boolean().default(false),
  privacy_consent: z.literal(true, {
    errorMap: () => ({
      message: "É necessário consentir com a política de privacidade",
    }),
  }),
});

export const titheFormSchema = titheFormBase.superRefine(requireCommunity);

export const createPaymentIntentSchema = titheFormBase
  .extend({ is_recurring: z.literal(false) })
  .superRefine(requireCommunity);

export const createRecurringSetupSchema = titheFormBase
  .extend({ is_recurring: z.literal(true) })
  .superRefine(requireCommunity);

export const createSubscriptionSchema = z.object({
  donor_id: z.string().uuid(),
  community_id: z.string().uuid().nullable().optional(),
  community_name: z.string().trim().max(160).optional().default(""),
  amount_cents: z.coerce
    .number()
    .int()
    .min(STRIPE_MIN_AMOUNT_CENTS)
    .max(10_000_000),
  stripe_customer_id: z.string().min(1),
  payment_method_id: z.string().min(1),
});

export const requestRecurringLinkSchema = z.object({
  email: emailLike(),
});

function emailLike() {
  return z.string().trim().toLowerCase().email("E-mail inválido").max(200);
}

export const confirmRecurringCancellationSchema = z.object({
  subscription_id: z.string().min(1),
  token: z.string().min(1),
});

export type TitheFormInput = z.infer<typeof titheFormSchema>;
export type CreateSubscriptionInput = z.infer<typeof createSubscriptionSchema>;