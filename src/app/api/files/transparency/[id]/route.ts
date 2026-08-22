import { NextResponse } from "next/server";
import { createAdminClient } from "@/lib/supabase/admin";
import { getTransparencyDocumentUrl } from "@/lib/storage/helpers";
import { jsonError } from "@/lib/http";

export const runtime = "nodejs";

export async function GET(
  _request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;

  const admin = createAdminClient();
  const { data: doc } = await admin
    .from("transparency_documents")
    .select("file_path, financial_periods(status)")
    .eq("id", id)
    .maybeSingle();

  if (!doc) {
    return jsonError("Documento não encontrado", 404);
  }

  const periodStatus = (doc.financial_periods as { status?: string } | null)
    ?.status;

  if (periodStatus !== "published") {
    return jsonError("Documento indisponível", 403);
  }

  const url = await getTransparencyDocumentUrl(doc.file_path);
  if (!url) {
    return jsonError("Falha ao gerar acesso ao documento", 500);
  }

  return NextResponse.redirect(url);
}