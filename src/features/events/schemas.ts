import { z } from "zod";

export const eventInputSchema = z.object({
  title: z.string().trim().min(3, "Informe o título").max(200),
  slug: z
    .string()
    .trim()
    .regex(/^[a-z0-9-]+$/, "Slug inválido")
    .optional(),
  description: z.string().trim().max(5000).optional().default(""),
  starts_at: z.string().min(1, "Informe a data de início"),
  ends_at: z.string().optional().nullable(),
  location: z.string().trim().max(300).optional().default(""),
  status: z.enum(["draft", "published", "archived"]).default("draft"),
});

export const eventStatusSchema = z.object({
  status: z.enum(["draft", "published", "archived"]),
});

export type EventInput = z.infer<typeof eventInputSchema>;