import { createAdminClient } from "@/lib/supabase/admin";

export interface DashboardSummary {
  total_collected_cents: number;
  total_expense_cents: number;
  balance_cents: number;
  total_donors: number;
  total_paid_contributions: number;
  pending_payments: number;
  failed_payments: number;
  canceled_recurring: number;
  community_ranking: {
    community_id: string | null;
    community_name: string;
    total_cents: number;
    count: number;
  }[];
  monthly_evolution: { month: string; total_cents: number }[];
}

export async function getDashboardSummary(): Promise<DashboardSummary> {
  const admin = createAdminClient();

  const [{ data: paidRows }, { count: totalDonors }, { count: pending }, { count: failed }, { count: canceledRecurring }] =
    await Promise.all([
      admin
        .from("contribution_payments")
        .select("amount_cents, paid_at, community_id, community_name, status")
        .eq("status", "paid"),
      admin.from("donors").select("id", { count: "exact", head: true }),
      admin
        .from("contribution_payments")
        .select("id", { count: "exact", head: true })
        .eq("status", "pending"),
      admin
        .from("contribution_payments")
        .select("id", { count: "exact", head: true })
        .eq("status", "failed"),
      admin
        .from("recurring_donations")
        .select("id", { count: "exact", head: true })
        .eq("status", "canceled"),
    ]);

  const paid = (paidRows ?? []) as {
    amount_cents: number;
    paid_at: string;
    community_id: string | null;
    community_name: string | null;
  }[];

  const totalCollected = paid.reduce((acc, r) => acc + r.amount_cents, 0);

  const { data: communities } = await admin
    .from("communities")
    .select("id, name");

  const communityMap = new Map(
    (communities ?? []).map((c) => [c.id, c.name as string])
  );

  const communityAgg = new Map<
    string,
    { total_cents: number; count: number; label: string }
  >();
  for (const row of paid) {
    const key = row.community_id ?? `nome:${row.community_name ?? ""}`;
    const label =
      (row.community_id ? communityMap.get(row.community_id) : null) ??
      (row.community_name?.trim() || "Sem comunidade");
    const cur = communityAgg.get(key) ?? {
      total_cents: 0,
      count: 0,
      label,
    };
    cur.total_cents += row.amount_cents;
    cur.count += 1;
    communityAgg.set(key, cur);
  }

  const communityRanking = [...communityAgg.entries()]
    .map(([key, v]) => ({
      community_id: key.startsWith("nome:") ? null : key,
      community_name: v.label,
      total_cents: v.total_cents,
      count: v.count,
    }))
    .sort((a, b) => b.total_cents - a.total_cents);

  const monthlyMap = new Map<string, number>();
  for (const row of paid) {
    if (!row.paid_at) continue;
    const month = row.paid_at.slice(0, 7);
    monthlyMap.set(month, (monthlyMap.get(month) ?? 0) + row.amount_cents);
  }
  const monthlyEvolution = [...monthlyMap.entries()]
    .sort((a, b) => a[0].localeCompare(b[0]))
    .map(([month, total_cents]) => ({ month, total_cents }));

  const { data: expenseRows } = await admin
    .from("financial_entries")
    .select("amount_cents")
    .eq("type", "expense")
    .eq("source", "manual");
  const totalExpense =
    (expenseRows ?? []).reduce((acc, r) => acc + r.amount_cents, 0) ?? 0;

  return {
    total_collected_cents: totalCollected,
    total_expense_cents: totalExpense,
    balance_cents: totalCollected - totalExpense,
    total_donors: totalDonors ?? 0,
    total_paid_contributions: paid.length,
    pending_payments: pending ?? 0,
    failed_payments: failed ?? 0,
    canceled_recurring: canceledRecurring ?? 0,
    community_ranking: communityRanking,
    monthly_evolution: monthlyEvolution,
  };
}