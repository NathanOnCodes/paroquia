"use client";

import { useState } from "react";
import Button from "@mui/material/Button";
import TextField from "@mui/material/TextField";
import Checkbox from "@mui/material/Checkbox";
import FormControlLabel from "@mui/material/FormControlLabel";
import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import CircularProgress from "@mui/material/CircularProgress";
import Link from "next/link";

export default function ContactForm() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [sent, setSent] = useState(false);
  const [form, setForm] = useState({
    name: "",
    email: "",
    phone: "",
    message: "",
    privacy_consent: false,
  });

  function update(field: keyof typeof form, value: string | boolean) {
    setForm((prev) => ({ ...prev, [field]: value }));
  }

  async function handleSubmit() {
    setError(null);
    if (!form.name || !form.email || !form.message) {
      setError("Preencha nome, e-mail e mensagem.");
      return;
    }
    if (!form.privacy_consent) {
      setError("É necessário consentir com a política de privacidade.");
      return;
    }

    setLoading(true);
    try {
      const res = await fetch("/api/contact", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(form),
      });
      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error ?? "Falha ao enviar mensagem");
      }
      setSent(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Erro inesperado");
    } finally {
      setLoading(false);
    }
  }

  if (sent) {
    return (
      <Alert severity="success">
        Mensagem enviada com sucesso. Em breve entraremos em contato.
      </Alert>
    );
  }

  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 2, maxWidth: 520 }}>
      {error && (
        <Alert severity="error" onClose={() => setError(null)}>
          {error}
        </Alert>
      )}
      <TextField
        label="Nome"
        required
        value={form.name}
        onChange={(e) => update("name", e.target.value)}
        fullWidth
      />
      <TextField
        label="E-mail"
        type="email"
        required
        value={form.email}
        onChange={(e) => update("email", e.target.value)}
        fullWidth
      />
      <TextField
        label="Telefone (opcional)"
        value={form.phone}
        onChange={(e) => update("phone", e.target.value)}
        fullWidth
      />
      <TextField
        label="Mensagem"
        required
        multiline
        rows={5}
        value={form.message}
        onChange={(e) => update("message", e.target.value)}
        fullWidth
      />
      <FormControlLabel
        control={
          <Checkbox
            checked={form.privacy_consent}
            onChange={(e) => update("privacy_consent", e.target.checked)}
          />
        }
        label={
          <span>
            Li e aceito a{" "}
            <Link href="/privacidade" style={{ color: "#1976d2" }}>
              política de privacidade
            </Link>
            .
          </span>
        }
      />
      <Button variant="contained" size="large" onClick={handleSubmit} disabled={loading}>
        {loading ? <CircularProgress size={22} color="inherit" /> : "Enviar mensagem"}
      </Button>
    </Box>
  );
}