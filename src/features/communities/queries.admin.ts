import { createAdminClient } from "@/lib/supabase/admin";
import type { Community } from "@/lib/db-types";

export async function listCommunitiesAdmin(): Promise<Community[]> {
  const admin = createAdminClient();
  const { data } = await admin
    .from("communities")
    .select("*")
    .order("name", { ascending: true });
  return (data as Community[]) ?? [];
}

export async function getCommunityById(id: string): Promise<Community | null> {
  const admin = createAdminClient();
  const { data } = await admin
    .from("communities")
    .select("*")
    .eq("id", id)
    .maybeSingle();
  return (data as Community | null) ?? null;
}

export interface CommunityWithSelect {
  id: string;
  name: string;
  description: string | null;
  city: string | null;
  is_active: boolean;
  communities?: { name: string } | null;
}

export async function listAllCommunities(): Promise<CommunityWithSelect[]> {
  const admin = createAdminClient();
  const { data } = await admin
    .from("communities")
    .select("*")
    .order("name", { ascending: true });
  return (data ?? []) as CommunityWithSelect[];
}