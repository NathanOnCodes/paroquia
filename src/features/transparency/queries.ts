import { createClient } from "@/lib/supabase/server";
import { createAdminClient } from "@/lib/supabase/admin";
import type {
  FinancialEntry,
  FinancialPeriod,
  TransparencyDocument,
} from "@/lib/db-types";

export interface PeriodWithTotals extends FinancialPeriod {
  total_income_cents: number;
  total_expense_cents: number;
  balance_cents: number;
  contributions_count: number;
}

export async function listPublishedPeriods(): Promise<PeriodWithTotals[]> {
  const supabase = await createClient();
  const { data: periods } = await supabase
    .from("financial_periods")
    .select("*")
    .eq("status", "published")
    .order("start_date", { ascending: false });

  const list = (periods as FinancialPeriod[]) ?? [];
  const result: PeriodWithTotals[] = [];

  for (const period of list) {
    result.push(await withTotals(period));
  }
  return result;
}

export async function getPublishedPeriodById(
  id: string
): Promise<{
  period: PeriodWithTotals;
  entries: FinancialEntry[];
  documents: TransparencyDocument[];
} | null> {
  const supabase = await createClient();
  const { data: period } = await supabase
    .from("financial_periods")
    .select("*")
    .eq("id", id)
    .eq("status", "published")
    .maybeSingle();

  if (!period) return null;

  const withTotalsRow = await withTotals(period as FinancialPeriod);

  const { data: entries } = await supabase
    .from("financial_entries")
    .select("*")
    .eq("financial_period_id", id)
    .order("entry_date", { ascending: true });

  const admin = createAdminClient();
  const { data: documents } = await admin
    .from("transparency_documents")
    .select("*")
    .eq("financial_period_id", id)
    .order("created_at", { ascending: false });

  return {
    period: withTotalsRow,
    entries: (entries as FinancialEntry[]) ?? [],
    documents: (documents as TransparencyDocument[]) ?? [],
  };
}

async function withTotals(period: FinancialPeriod): Promise<PeriodWithTotals> {
  const supabase = await createClient();
  const { data: entries } = await supabase
    .from("financial_entries")
    .select("type, amount_cents")
    .eq("financial_period_id", period.id);

  let income = 0;
  let expense = 0;
  for (const entry of entries ?? []) {
    if (entry.type === "income") income += entry.amount_cents;
    else expense += entry.amount_cents;
  }

  const admin = createAdminClient();
  const { count } = await admin
    .from("contribution_payments")
    .select("id", { count: "exact", head: true })
    .eq("status", "paid")
    .gte("paid_at", new Date(`${period.start_date}T00:00:00`).toISOString())
    .lte("paid_at", new Date(`${period.end_date}T23:59:59`).toISOString());

  return {
    ...period,
    total_income_cents: income,
    total_expense_cents: expense,
    balance_cents: income - expense,
    contributions_count: count ?? 0,
  };
}