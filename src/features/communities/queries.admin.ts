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