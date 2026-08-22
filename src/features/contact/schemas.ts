import { z } from "zod";
import { emailSchema } from "@/features/donors/schemas";

export const contactInputSchema = z.object({
  name: z.string().trim().min(3, "Informe seu nome").max(200),
  email: emailSchema,
  phone: z
    .string()
    .trim()
    .regex(/^[0-9()\s-]{8,20}$/, "Telefone inválido")
    .optional()
    .or(z.literal("")),
  message: z.string().trim().min(10, "Mensagem muito curta").max(4000),
  privacy_consent: z.literal(true, {
    errorMap: () => ({ message: "É necessário consentir com a política de privacidade" }),
  }),
});

export const contactStatusSchema = z.object({
  status: z.enum(["new", "in_progress", "resolved", "archived"]),
});

export type ContactInput = z.infer<typeof contactInputSchema>;