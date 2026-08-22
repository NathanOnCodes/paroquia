import { createAdminClient } from "@/lib/supabase/admin";
import type { TransparencyDocument } from "@/lib/db-types";

export async function listDocumentsForPeriod(
  periodId: string
): Promise<TransparencyDocument[]> {
  const admin = createAdminClient();
  const { data } = await admin
    .from("transparency_documents")
    .select("*")
    .eq("financial_period_id", periodId)
    .order("created_at", { ascending: false });
  return (data as TransparencyDocument[]) ?? [];
}