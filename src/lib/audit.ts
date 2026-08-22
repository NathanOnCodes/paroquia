import { createAdminClient } from "@/lib/supabase/admin";

export async function logAudit(input: {
  actor_id: string;
  action: string;
  entity_type: string;
  entity_id?: string | null;
  metadata?: Record<string, unknown> | null;
}): Promise<void> {
  const supabase = createAdminClient();
  await supabase.from("audit_logs").insert({
    actor_id: input.actor_id,
    action: input.action,
    entity_type: input.entity_type,
    entity_id: input.entity_id ?? null,
    metadata: input.metadata ?? null,
  });
}