import { NextResponse } from "next/server";

export function jsonError(message: string, status: number) {
  return NextResponse.json({ error: message }, { status });
}

export function handleRouteError(error: unknown): NextResponse {
  if (error instanceof Error) {
    return jsonError(error.message, 400);
  }
  return jsonError("Erro inesperado", 500);
}

export function assertMethod(request: Request, methods: string[]): boolean {
  return methods.includes(request.method.toUpperCase());
}