"use client";

import { useState, useRef } from "react";
import { useRouter } from "next/navigation";
import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";

interface ActionResult {
  error?: string;
  ok?: boolean;
}

interface ServerFormProps {
  action: (formData: FormData) => Promise<ActionResult>;
  children: React.ReactNode;
  submitLabel?: string;
  cancelHref?: string;
  successMessage?: string;
}

export default function ServerForm({
  action,
  children,
  submitLabel = "Salvar",
  cancelHref,
  successMessage = "Salvo com sucesso!",
}: ServerFormProps) {
  const router = useRouter();
  const ref = useRef<HTMLFormElement>(null);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(formData: FormData) {
    setLoading(true);
    setError(null);
    setSuccess(null);
    try {
      const result = await action(formData);
      if (result?.error) {
        setError(result.error);
      } else {
        setSuccess(successMessage);
        router.refresh();
        ref.current?.reset();
      }
    } catch {
      setError("Ocorreu um erro ao processar a solicitação.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <form ref={ref} action={handleSubmit}>
      <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
        {error && (
          <Alert severity="error" onClose={() => setError(null)}>
            {error}
          </Alert>
        )}
        {success && (
          <Alert severity="success" onClose={() => setSuccess(null)}>
            {success}
          </Alert>
        )}
        {children}
        <Box sx={{ display: "flex", gap: 2 }}>
          <Button type="submit" variant="contained" disabled={loading}>
            {loading ? <CircularProgress size={20} color="inherit" /> : submitLabel}
          </Button>
          {cancelHref && (
            <Button component="a" href={cancelHref} disabled={loading}>
              Cancelar
            </Button>
          )}
        </Box>
      </Box>
    </form>
  );
}