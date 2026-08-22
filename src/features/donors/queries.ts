import { createAdminClient } from "@/lib/supabase/admin";
import type { Donor } from "@/lib/db-types";

export async function listDonors(search?: string): Promise<Donor[]> {
  const admin = createAdminClient();
  let query = admin
    .from("donors")
    .select("*, communities(name)")
    .order("created_at", { ascending: false })
    .limit(200);

  if (search && search.trim().length > 0) {
    query = query.or(
      `full_name.ilike.%${search}%,email.ilike.%${search}%,phone.ilike.%${search}%`
    );
  }

  const { data } = await query;
  return (data ?? []) as Donor[];
}

export async function getDonorById(id: string): Promise<Donor | null> {
  const admin = createAdminClient();
  const { data } = await admin
    .from("donors")
    .select("*, communities(name)")
    .eq("id", id)
    .maybeSingle();
  return (data as Donor | null) ?? null;
}

export async function listAllCommunities(): Promise<CommunityWithSelect[]> {
  const admin = createAdminClient();
  const { data } = await admin
    .from("communities")
    .select("*")
    .order("name", { ascending: true });
  return (data ?? []) as CommunityWithSelect[];
}

export interface CommunityWithSelect {
  id: string;
  name: string;
  description: string | null;
  city: string | null;
  is_active: boolean;
  communities?: { name: string } | null;
}