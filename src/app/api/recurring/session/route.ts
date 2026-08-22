import { NextResponse } from "next/server";
import { cookies } from "next/headers";
import {
  consumeMagicToken,
  validateMagicToken,
} from "@/features/recurring/services";

export const runtime = "nodejs";

const SESSION_MAX_AGE_SECONDS = 60 * 60;

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const token = searchParams.get("token");

  const baseUrl = new URL(
    process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000"
  );
  const home = `${baseUrl}/dizimo/gerenciar`;

  if (!token) {
    return NextResponse.redirect(home);
  }

  const record = await validateMagicToken(token);
  if (!record || record.consumed_at) {
    const url = new URL(home);
    url.searchParams.set("error", "invalid");
    return NextResponse.redirect(url);
  }

  await consumeMagicToken(record.id);

  const cookieStore = await cookies();
  cookieStore.set("recurring_mgmt", token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: SESSION_MAX_AGE_SECONDS,
  });

  return NextResponse.redirect(home);
}