import { createAdminClient } from "@/lib/supabase/admin";
import type { Profile, FinancialPeriod, FinancialEntry } from "@/lib/db-types";

export async function listStaffUsers(): Promise<Profile[]> {
  const admin = createAdminClient();
  const { data } = await admin
    .from("profiles")
    .select("*")
    .order("created_at", { ascending: true });
  return (data as Profile[]) ?? [];
}

export async function getStaffUserById(id: string): Promise<Profile | null> {
  const admin = createAdminClient();
  const { data } = await admin
    .from("profiles")
    .select("*")
    .eq("id", id)
    .maybeSingle();
  return (data as Profile | null) ?? null;
}

export async function listFinancialPeriodsAdmin(): Promise<FinancialPeriod[]> {
  const admin = createAdminClient();
  const { data } = await admin
    .from("financial_periods")
    .select("*")
    .order("start_date", { ascending: false });
  return (data as FinancialPeriod[]) ?? [];
}

export async function getFinancialPeriodById(
  id: string
): Promise<{
  period: FinancialPeriod;
  entries: FinancialEntry[];
} | null> {
  const admin = createAdminClient();
  const { data: period } = await admin
    .from("financial_periods")
    .select("*")
    .eq("id", id)
    .maybeSingle();
  if (!period) return null;

  const { data: entries } = await admin
    .from("financial_entries")
    .select("*")
    .eq("financial_period_id", id)
    .order("entry_date", { ascending: false });

  return { period: period as FinancialPeriod, entries: (entries as FinancialEntry[]) ?? [] };
}