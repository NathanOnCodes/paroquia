import { createAdminClient } from "@/lib/supabase/admin";
import type { Profile } from "@/lib/db-types";

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