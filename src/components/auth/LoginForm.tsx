"use client";

import { useState } from "react";
import Button from "@mui/material/Button";
import TextField from "@mui/material/TextField";
import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import CircularProgress from "@mui/material/CircularProgress";
import { loginAction } from "@/features/auth/actions";

export default function LoginForm() {
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(formData: FormData) {
    setLoading(true);
    setError(null);
    const result = await loginAction(formData);
    if (result?.error) {
      setError(result.error);
      setLoading(false);
    }
  }

  return (
    <Box component="form" action={handleSubmit} sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
      {error && <Alert severity="error">{error}</Alert>}
      <TextField
        label="E-mail"
        name="email"
        type="email"
        required
        autoComplete="email"
        fullWidth
      />
      <TextField
        label="Senha"
        name="password"
        type="password"
        required
        autoComplete="current-password"
        fullWidth
      />
      <Button type="submit" variant="contained" size="large" disabled={loading}>
        {loading ? <CircularProgress size={22} color="inherit" /> : "Entrar"}
      </Button>
    </Box>
  );
}