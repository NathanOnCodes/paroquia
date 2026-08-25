"use client";

import { useState } from "react";
import Button from "@mui/material/Button";
import TextField from "@mui/material/TextField";
import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import CircularProgress from "@mui/material/CircularProgress";

export default function RecurringRequestForm() {
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  async function handleSubmit() {
    setError(null);
    setMessage(null);
    if (!email) {
      setError("Informe seu e-mail.");
      return;
    }
    setLoading(true);
    try {
      const res = await fetch("/api/recurring/request-link", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email }),
      });
      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error ?? "Falha ao solicitar link");
      }
      setMessage(data.message ?? "Verifique seu e-mail.");
      setEmail("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Erro inesperado");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 2, maxWidth: 440 }}>
      {error && (
        <Alert severity="error" onClose={() => setError(null)}>
          {error}
        </Alert>
      )}
      {message && <Alert severity="success">{message}</Alert>}
      <TextField
        label="E-mail cadastrado"
        type="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        fullWidth
        helperText="Enviaremos um link de acesso seguro para gerenciar suas recorrências."
      />
      <Button variant="contained" size="large" onClick={handleSubmit} disabled={loading}>
        {loading ? <CircularProgress size={22} color="inherit" /> : "Receber link de acesso"}
      </Button>
    </Box>
  );
}