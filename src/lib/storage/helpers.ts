import { createAdminClient } from "@/lib/supabase/admin";
import { env } from "@/lib/env";

export function getEventImageUrl(imagePath: string | null): string | null {
  if (!imagePath) return null;
  return `${env.NEXT_PUBLIC_SUPABASE_URL}/storage/v1/object/public/events/${imagePath}`;
}

export async function getTransparencyDocumentUrl(
  filePath: string,
  expiresInSeconds = 60
): Promise<string | null> {
  const admin = createAdminClient();
  const { data, error } = await admin.storage
    .from("transparency")
    .createSignedUrl(filePath, expiresInSeconds);
  if (error) return null;
  return data.signedUrl;
}