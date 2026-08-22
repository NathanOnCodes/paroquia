"use client";

import { useEffect } from "react";

/**
 * Error boundary global. Substitui o root layout inteiro quando um erro
 * acontece nele, por isso define seus próprios <html> e <body>.
 */
export default function GlobalError({
  error,
  retry,
}: {
  error: Error & { digest?: string };
  retry: () => void;
}) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <html lang="pt-BR">
      <body
        style={{
          margin: 0,
          minHeight: "100vh",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          backgroundColor: "#faf9f7",
          color: "#1c1c1c",
          fontFamily:
            "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Arial, sans-serif",
        }}
      >
        <div style={{ textAlign: "center", padding: "2rem", maxWidth: 480 }}>
          <div
            style={{
              fontSize: "3rem",
              fontWeight: 700,
              color: "#1a3a5c",
              marginBottom: "0.5rem",
            }}
          >
            Erro
          </div>
          <h1 style={{ fontSize: "1.5rem", margin: "0 0 0.75rem" }}>
            Algo deu errado
          </h1>
          <p style={{ color: "#555", margin: "0 0 1.5rem", lineHeight: 1.6 }}>
            Ocorreu um erro inesperado. Tente novamente ou volte mais tarde.
          </p>
          {error.digest && (
            <p
              style={{
                fontSize: "0.8rem",
                color: "#b45309",
                backgroundColor: "#fef3c7",
                borderRadius: 8,
                padding: "0.5rem 0.75rem",
                margin: "0 0 1.5rem",
              }}
            >
              Código do erro: {error.digest}
            </p>
          )}
          <button
            onClick={retry}
            style={{
              backgroundColor: "#1a3a5c",
              color: "#fff",
              border: "none",
              borderRadius: 8,
              padding: "0.75rem 1.5rem",
              fontSize: "1rem",
              fontWeight: 600,
              cursor: "pointer",
            }}
          >
            Tentar novamente
          </button>
        </div>
      </body>
    </html>
  );
}