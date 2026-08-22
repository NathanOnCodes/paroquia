import { createAdminClient } from "@/lib/supabase/admin";
import type { Donor } from "@/lib/db-types";
import type { DonorInput } from "@/features/donors/schemas";

export async function findDonorByEmail(email: string): Promise<Donor | null> {
  const supabase = createAdminClient();
  const { data } = await supabase
    .from("donors")
    .select("*")
    .eq("email", email.toLowerCase())
    .maybeSingle();
  return (data as Donor | null) ?? null;
}

export async function upsertDonor(input: DonorInput): Promise<Donor> {
  const supabase = createAdminClient();
  const existing = await findDonorByEmail(input.email);

  if (existing) {
    const { data, error } = await supabase
      .from("donors")
      .update({
        full_name: input.full_name,
        phone: input.phone || null,
        city: input.city,
        neighborhood: input.neighborhood,
        community_id: input.community_id ?? null,
        community_name: input.community_id ? null : input.community_name || null,
      })
      .eq("id", existing.id)
      .select("*")
      .single();
    if (error) throw new Error("Falha ao atualizar dizimista");
    return data as Donor;
  }

  const { data, error } = await supabase
    .from("donors")
    .insert({
      full_name: input.full_name,
      email: input.email.toLowerCase(),
      phone: input.phone || null,
      city: input.city,
      neighborhood: input.neighborhood,
      community_id: input.community_id ?? null,
      community_name: input.community_id ? null : input.community_name || null,
      privacy_consent_at: new Date().toISOString(),
    })
    .select("*")
    .single();
  if (error) throw new Error("Falha ao cadastrar dizimista");
  return data as Donor;
}

export async function ensureDonorConsent(donorId: string): Promise<void> {
  const supabase = createAdminClient();
  await supabase
    .from("donors")
    .update({ privacy_consent_at: new Date().toISOString() })
    .eq("id", donorId);
}