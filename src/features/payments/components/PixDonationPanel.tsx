"use client";

import { useState } from "react";
import Image from "next/image";
import Alert from "@mui/material/Alert";
import Button from "@mui/material/Button";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import ContentCopyIcon from "@mui/icons-material/ContentCopy";
import EmptyState from "@/design-system/feedback/EmptyState";

export default function PixDonationPanel({
  payment,
}: {
  payment: {
    brCode: string;
    qrCodeImage: string;
    receiverName: string;
    receiverCity: string;
  } | null;
}) {
  const [copied, setCopied] = useState(false);

  if (!payment) {
    return (
      <EmptyState
        title="Pix ainda não configurado"
        description="A administração ainda não cadastrou uma chave Pix para esta paróquia."
      />
    );
  }

  const currentPayment = payment;

  async function copyCode() {
    await navigator.clipboard.writeText(currentPayment.brCode);
    setCopied(true);
    window.setTimeout(() => setCopied(false), 2500);
  }

  return (
    <Card sx={{ maxWidth: 560, mx: "auto" }}>
      <CardContent>
        <Stack spacing={2} sx={{ alignItems: "center" }}>
          <Typography variant="h6" component="h2" sx={{ textAlign: "center" }}>
            Faça sua contribuição via Pix
          </Typography>
          <Typography color="text.secondary" sx={{ textAlign: "center" }}>
            Escaneie o QR Code e informe o valor que deseja contribuir.
          </Typography>
          <Image
            src={currentPayment.qrCodeImage}
            alt={`QR Code Pix para ${currentPayment.receiverName}, ${currentPayment.receiverCity}`}
            width={240}
            height={240}
            unoptimized
            style={{ display: "block", maxWidth: "100%", height: "auto" }}
          />
          <Typography variant="body2" color="text.secondary" sx={{ textAlign: "center" }}>
            Recebedor: {currentPayment.receiverName}
          </Typography>
          <Button onClick={copyCode} variant="outlined" startIcon={<ContentCopyIcon />}>
            Copiar Pix copia e cola
          </Button>
          {copied && (
            <Alert severity="success" role="status">
              Pix copia e cola copiado.
            </Alert>
          )}
        </Stack>
      </CardContent>
    </Card>
  );
}
