import { createClient } from "@/lib/supabase/server";
import type { Community } from "@/lib/db-types";

export async function listActiveCommunities(): Promise<Community[]> {
  const supabase = await createClient();
  const { data } = await supabase
    .from("communities")
    .select("*")
    .eq("is_active", true)
    .order("name", { ascending: true });
  return (data as Community[]) ?? [];
}