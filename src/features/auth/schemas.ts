import { z } from "zod";

export const loginSchema = z.object({
  email: z.string().trim().toLowerCase().email("E-mail inválido").max(200),
  password: z.string().min(8, "A senha deve ter no mínimo 8 caracteres"),
});

export const createStaffUserSchema = z.object({
  email: z.string().trim().toLowerCase().email("E-mail inválido").max(200),
  full_name: z.string().trim().min(3, "Informe o nome completo").max(200),
  password: z.string().min(8, "A senha deve ter no mínimo 8 caracteres"),
  role: z.enum(["admin", "assistente"]),
  is_active: z.boolean().default(true),
});

export const updateStaffUserSchema = z.object({
  full_name: z.string().trim().min(3, "Informe o nome completo").max(200),
  role: z.enum(["admin", "assistente"]),
  is_active: z.boolean(),
});

export type LoginInput = z.infer<typeof loginSchema>;
export type CreateStaffUserInput = z.infer<typeof createStaffUserSchema>;