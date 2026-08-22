import { createAdminClient } from "@/lib/supabase/admin";
import type { ContributionPayment } from "@/lib/db-types";

export interface ContributionRow extends ContributionPayment {
  donors?: {
    full_name: string;
    email: string;
    community_name: string | null;
  } | null;
  communities?: { name: string } | null;
}

export async function listContributionsAdmin(): Promise<ContributionRow[]> {
  const admin = createAdminClient();
  const { data } = await admin
    .from("contribution_payments")
    .select("*, donors(full_name, email, community_name), communities(name)")
    .order("created_at", { ascending: false })
    .limit(300);
  return (data ?? []) as ContributionRow[];
}

export async function listContributionsByDonor(
  donorId: string
): Promise<ContributionRow[]> {
  const admin = createAdminClient();
  const { data } = await admin
    .from("contribution_payments")
    .select("*, donors(full_name, email, community_name), communities(name)")
    .eq("donor_id", donorId)
    .order("created_at", { ascending: false });
  return (data ?? []) as ContributionRow[];
}