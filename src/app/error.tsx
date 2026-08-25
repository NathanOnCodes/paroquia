"use client";

import { useEffect } from "react";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";
import Alert from "@mui/material/Alert";

export default function Error({
  error,
  retry,
}: {
  error: Error & { digest?: string };
  retry: () => void;
}) {
  useEffect(() => {
    if (process.env.NODE_ENV !== "production") console.error(error);
  }, [error]);

  return (
    <Box
      sx={{
        minHeight: "60vh",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        textAlign: "center",
        gap: 2,
        py: 8,
      }}
    >
      <Typography variant="h4" component="h1">
        Algo deu errado
      </Typography>
      <Typography color="text.secondary" sx={{ maxWidth: 420 }}>
        Ocorreu um erro inesperado ao carregar esta página. Tente novamente ou
        volte para o início.
      </Typography>
      {process.env.NODE_ENV !== "production" && error.digest && (
        <Alert severity="warning" sx={{ maxWidth: 420 }}>
          Código do erro: {error.digest}
        </Alert>
      )}
      <Box sx={{ display: "flex", gap: 2, mt: 2 }}>
        <Button variant="contained" onClick={retry}>
          Tentar novamente
        </Button>
        <Button href="/" variant="outlined">
          Voltar ao início
        </Button>
      </Box>
    </Box>
  );
}
