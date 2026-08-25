import { createClient } from "@/lib/supabase/server";
import { createAdminClient } from "@/lib/supabase/admin";
import type { ChurchEvent } from "@/lib/db-types";
import { env } from "@/lib/env";

export async function listPublishedEvents(): Promise<ChurchEvent[]> {
  if (env.LOCAL_DEMO_MODE) return [];

  const supabase = await createClient();
  const { data } = await supabase
    .from("events")
    .select("*")
    .eq("status", "published")
    .gte("starts_at", new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString())
    .order("starts_at", { ascending: true })
    .limit(50);
  return (data as ChurchEvent[]) ?? [];
}

export async function getPublishedEventBySlug(slug: string): Promise<ChurchEvent | null> {
  const supabase = await createClient();
  const { data } = await supabase
    .from("events")
    .select("*")
    .eq("slug", slug)
    .eq("status", "published")
    .maybeSingle();
  return (data as ChurchEvent | null) ?? null;
}

export async function listEventsAdmin(): Promise<ChurchEvent[]> {
  const admin = createAdminClient();
  const { data } = await admin
    .from("events")
    .select("*")
    .order("starts_at", { ascending: false })
    .limit(200);
  return (data as ChurchEvent[]) ?? [];
}

export async function getEventById(id: string): Promise<ChurchEvent | null> {
  const admin = createAdminClient();
  const { data } = await admin
    .from("events")
    .select("*")
    .eq("id", id)
    .maybeSingle();
  return (data as ChurchEvent | null) ?? null;
}
