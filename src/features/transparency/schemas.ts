import { z } from "zod";

export const financialPeriodSchema = z.object({
  name: z.string().trim().min(3, "Informe o nome do período").max(200),
  start_date: z.string().min(1, "Informe a data inicial"),
  end_date: z.string().min(1, "Informe a data final"),
  status: z.enum(["draft", "review", "published", "archived"]).default("draft"),
});

export const financialEntrySchema = z.object({
  financial_period_id: z.string().uuid("Período inválido"),
  type: z.enum(["income", "expense"]),
  category: z.string().trim().min(2, "Informe a categoria").max(120),
  description: z.string().trim().min(3, "Informe a descrição").max(500),
  amount_cents: z.coerce
    .number()
    .int("Valor inválido")
    .min(1, "Valor deve ser maior que zero")
    .max(1_000_000_000),
  entry_date: z.string().min(1, "Informe a data do lançamento"),
});

export const transparencyStatusSchema = z.object({
  status: z.enum(["draft", "review", "published", "archived"]),
});

export type FinancialPeriodInput = z.infer<typeof financialPeriodSchema>;
export type FinancialEntryInput = z.infer<typeof financialEntrySchema>;