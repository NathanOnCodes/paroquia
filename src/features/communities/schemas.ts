import { z } from "zod";

export const communitySchema = z.object({
  name: z.string().trim().min(3, "Informe o nome da comunidade").max(120),
  description: z.string().trim().max(500).optional().default(""),
  city: z.string().trim().max(120).optional().default(""),
  is_active: z.boolean().default(true),
});

export const communityStatusSchema = z.object({
  is_active: z.boolean(),
});

export type CommunityInput = z.infer<typeof communitySchema>;