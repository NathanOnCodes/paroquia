import { createAdminClient } from "@/lib/supabase/admin";
import type { ContactMessage, Profile } from "@/lib/db-types";

export interface ContactRow extends ContactMessage {
  profiles?: Pick<Profile, "full_name"> | null;
}

export async function listContactsAdmin(): Promise<ContactRow[]> {
  const admin = createAdminClient();
  const { data } = await admin
    .from("contact_messages")
    .select("*, profiles(full_name)")
    .order("created_at", { ascending: false })
    .limit(200);
  return (data ?? []) as ContactRow[];
}