import { z } from "zod";

export const emailSchema = z
  .string()
  .trim()
  .toLowerCase()
  .email("E-mail inválido")
  .max(200);

export const phoneSchema = z
  .string()
  .trim()
  .regex(/^[0-9()\s-]{8,20}$/, "Telefone inválido")
  .optional()
  .or(z.literal(""));

export const donorInputSchema = z.object({
  full_name: z.string().trim().min(3, "Informe o nome completo").max(200),
  email: emailSchema,
  phone: phoneSchema,
  city: z.string().trim().max(120).optional().default(""),
  neighborhood: z.string().trim().max(120).optional().default(""),
  community_id: z.string().uuid("Comunidade inválida").optional().nullable(),
  community_name: z
    .string()
    .trim()
    .max(160, "Nome muito longo")
    .optional()
    .default(""),
});

export const donorStatusSchema = z.object({
  is_active: z.boolean(),
});

export type DonorInput = z.infer<typeof donorInputSchema>;