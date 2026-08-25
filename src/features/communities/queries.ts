import { createClient } from "@/lib/supabase/server";
import type { Community } from "@/lib/db-types";
import { env } from "@/lib/env";

export async function listActiveCommunities(): Promise<Community[]> {
  if (env.LOCAL_DEMO_MODE) return [];

  const supabase = await createClient();
  const { data } = await supabase
    .from("communities")
    .select("*")
    .eq("is_active", true)
    .order("name", { ascending: true });
  return (data as Community[]) ?? [];
}
