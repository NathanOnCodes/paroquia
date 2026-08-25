import { createClient } from "@/lib/supabase/server";
import { DEFAULT_SITE_SETTINGS, type SiteSettings } from "@/features/settings/types";
import { env } from "@/lib/env";

export async function getSiteSettings(): Promise<SiteSettings> {
  if (env.LOCAL_DEMO_MODE) return DEFAULT_SITE_SETTINGS;

  try {
    const supabase = await createClient();
    const { data } = await supabase
      .from("site_settings")
      .select("*")
      .eq("id", 1)
      .maybeSingle();

    if (!data) return DEFAULT_SITE_SETTINGS;
    return { ...DEFAULT_SITE_SETTINGS, ...(data as Partial<SiteSettings>) };
  } catch {
    return DEFAULT_SITE_SETTINGS;
  }
}
