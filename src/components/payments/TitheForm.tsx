"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
  loadStripe,
  type Stripe,
  type StripeElements,
} from "@stripe/stripe-js";
import Button from "@mui/material/Button";
import TextField from "@mui/material/TextField";
import MenuItem from "@mui/material/MenuItem";
import Checkbox from "@mui/material/Checkbox";
import FormControlLabel from "@mui/material/FormControlLabel";
import Alert from "@mui/material/Alert";
import AlertTitle from "@mui/material/AlertTitle";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import CircularProgress from "@mui/material/CircularProgress";
import Link from "next/link";
import { env } from "@/lib/env";
import type { Community } from "@/lib/db-types";

const stripePromise = loadStripe(env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY);

interface TitheFormProps {
  communities: Community[];
}

interface FormState {
  full_name: string;
  email: string;
  phone: string;
  community_id: string;
  community_name: string;
  city: string;
  neighborhood: string;
  amount_reais: string;
  is_recurring: boolean;
  privacy_consent: boolean;
}

const initialState: FormState = {
  full_name: "",
  email: "",
  phone: "",
  community_id: "",
  community_name: "",
  city: "",
  neighborhood: "",
  amount_reais: "",
  is_recurring: false,
  privacy_consent: false,
};

export default function TitheForm({ communities }: TitheFormProps) {
  const router = useRouter();
  const [form, setForm] = useState<FormState>(initialState);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [stripe, setStripe] = useState<Stripe | null>(null);
  const [elements, setElements] = useState<StripeElements | null>(null);
  const [clientSecret, setClientSecret] = useState<string | null>(null);
  const [stage, setStage] = useState<"form" | "payment">("form");
  const [pendingIntent, setPendingIntent] = useState<{
    type: "one_time" | "recurring";
    donor_id: string;
    stripe_customer_id?: string;
    community_id: string | null;
    community_name: string;
    amount_cents: number;
  } | null>(null);
  const [processing, setProcessing] = useState(false);

  useEffect(() => {
    if (!clientSecret) return;
    let cancelled = false;

    (async () => {
      const stripeClient = await stripePromise;
      if (!stripeClient || cancelled) return;
      setStripe(stripeClient);

      const elementsInstance = stripeClient.elements({
        clientSecret,
        appearance: { theme: "stripe" },
      });
      setElements(elementsInstance);

      const paymentElement = elementsInstance.create("payment");
      paymentElement.mount("#payment-element");
    })();

    return () => {
      cancelled = true;
    };
  }, [clientSecret]);

  function updateField(field: keyof FormState, value: string | boolean) {
    setForm((prev) => ({ ...prev, [field]: value }));
  }

  function parseAmountToCents(reais: string): number | null {
    const cleaned = reais.replace("R$", "").replace(/\s/g, "").replace(".", "").replace(",", ".");
    const value = Number(cleaned);
    if (!Number.isFinite(value) || value <= 0) return null;
    return Math.round(value * 100);
  }

  function isCustomCommunity(): boolean {
    return form.community_id === "other";
  }

  async function handleNext() {
    setError(null);
    const amountCents = parseAmountToCents(form.amount_reais);
    if (!amountCents) {
      setError("Informe um valor válido (ex.: 20,00)");
      return;
    }
    if (!form.full_name || !form.email || !form.phone || !form.city) {
      setError("Preencha todos os campos obrigatórios.");
      return;
    }
    if (isCustomCommunity() && !form.community_name.trim()) {
      setError("Informe o nome da sua comunidade.");
      return;
    }
    if (!isCustomCommunity() && !form.community_id) {
      setError("Selecione a sua comunidade.");
      return;
    }
    if (!form.privacy_consent) {
      setError("É necessário consentir com a política de privacidade.");
      return;
    }
    if (amountCents < 100) {
      setError("O valor mínimo é R$ 1,00.");
      return;
    }

    const communityId = isCustomCommunity() ? null : form.community_id;
    const communityName = isCustomCommunity()
      ? form.community_name.trim()
      : "";

    setLoading(true);
    try {
      const payload = {
        full_name: form.full_name,
        email: form.email,
        phone: form.phone,
        community_id: communityId,
        community_name: communityName,
        city: form.city,
        neighborhood: form.neighborhood,
        amount_cents: amountCents,
        is_recurring: form.is_recurring,
        privacy_consent: true,
      };

      if (form.is_recurring) {
        const res = await fetch("/api/payments/create-recurring", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error ?? "Falha ao iniciar recorrência");
        setPendingIntent({
          type: "recurring",
          donor_id: data.donor_id,
          stripe_customer_id: data.stripe_customer_id,
          community_id: communityId,
          community_name: communityName,
          amount_cents: amountCents,
        });
        setClientSecret(data.client_secret);
      } else {
        const res = await fetch("/api/payments/create-payment", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error ?? "Falha ao iniciar pagamento");
        setPendingIntent({
          type: "one_time",
          donor_id: data.donor_id ?? "",
          community_id: communityId,
          community_name: communityName,
          amount_cents: amountCents,
        });
        setClientSecret(data.client_secret);
      }
      setStage("payment");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Erro inesperado");
    } finally {
      setLoading(false);
    }
  }

  async function handlePay() {
    if (!stripe || !elements || !pendingIntent) return;
    setProcessing(true);
    setError(null);

    const returnUrl = `${window.location.origin}/dizimo/sucesso`;

    try {
      if (pendingIntent.type === "one_time") {
        const result = await stripe.confirmPayment({
          elements,
          confirmParams: { return_url: returnUrl },
          redirect: "if_required",
        });
        if (result.error) {
          setError(result.error.message ?? "Pagamento não concluído");
          return;
        }
        if (result.paymentIntent?.status === "succeeded") {
          router.push("/dizimo/sucesso");
          return;
        }
        router.push(returnUrl);
      } else {
        const result = await stripe.confirmSetup({
          elements,
          confirmParams: { return_url: returnUrl },
          redirect: "if_required",
        });
        if (result.error) {
          setError(result.error.message ?? "Não foi possível salvar o cartão");
          return;
        }
        if (result.setupIntent?.status !== "succeeded" || !result.setupIntent.payment_method) {
          setError("Não foi possível confirmar o pagamento.");
          return;
        }

        const subRes = await fetch("/api/payments/create-subscription", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            donor_id: pendingIntent.donor_id,
            community_id: pendingIntent.community_id,
            community_name: pendingIntent.community_name,
            amount_cents: pendingIntent.amount_cents,
            stripe_customer_id: pendingIntent.stripe_customer_id,
            payment_method_id: result.setupIntent.payment_method,
          }),
        });
        const subData = await subRes.json();
        if (!subRes.ok) {
          throw new Error(subData.error ?? "Falha ao criar recorrência");
        }
        router.push("/dizimo/sucesso");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Erro ao processar pagamento");
    } finally {
      setProcessing(false);
    }
  }

  return (
    <Box sx={{ maxWidth: 520, mx: "auto" }}>
      {stage === "form" ? (
        <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
          {error && (
            <Alert severity="error" onClose={() => setError(null)}>
              {error}
            </Alert>
          )}
          <TextField
            label="Nome completo"
            required
            value={form.full_name}
            onChange={(e) => updateField("full_name", e.target.value)}
            fullWidth
          />
          <TextField
            label="E-mail"
            type="email"
            required
            value={form.email}
            onChange={(e) => updateField("email", e.target.value)}
            fullWidth
          />
          <TextField
            label="Telefone"
            required
            value={form.phone}
            onChange={(e) => updateField("phone", e.target.value)}
            fullWidth
            placeholder="(00) 00000-0000"
          />
          <TextField
            select
            label="Comunidade"
            required
            value={form.community_id}
            onChange={(e) => updateField("community_id", e.target.value)}
            fullWidth
          >
            {communities.map((c) => (
              <MenuItem key={c.id} value={c.id}>
                {c.name}
              </MenuItem>
            ))}
            <MenuItem value="other">Minha comunidade não está na lista</MenuItem>
          </TextField>
          {isCustomCommunity() && (
            <TextField
              label="Qual é a sua comunidade?"
              required
              value={form.community_name}
              onChange={(e) => updateField("community_name", e.target.value)}
              fullWidth
              helperText="Escreva o nome da sua comunidade ou paróquia, mesmo que seja de outra cidade."
            />
          )}
          <Box sx={{ display: "flex", gap: 2, flexWrap: "wrap" }}>
            <TextField
              label="Cidade"
              required
              value={form.city}
              onChange={(e) => updateField("city", e.target.value)}
              sx={{ flex: 1, minWidth: 180 }}
            />
            <TextField
              label="Bairro (opcional)"
              value={form.neighborhood}
              onChange={(e) => updateField("neighborhood", e.target.value)}
              sx={{ flex: 1, minWidth: 180 }}
            />
          </Box>
          <TextField
            label="Valor do dízimo (R$)"
            required
            value={form.amount_reais}
            onChange={(e) => updateField("amount_reais", e.target.value)}
            fullWidth
            placeholder="20,00"
            slotProps={{ htmlInput: { inputMode: "decimal" } }}
          />
          <FormControlLabel
            control={
              <Checkbox
                checked={form.is_recurring}
                onChange={(e) => updateField("is_recurring", e.target.checked)}
              />
            }
            label="Quero que o dízimo seja recorrente todo mês"
          />
          <FormControlLabel
            control={
              <Checkbox
                checked={form.privacy_consent}
                onChange={(e) => updateField("privacy_consent", e.target.checked)}
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
          <Button
            variant="contained"
            size="large"
            onClick={handleNext}
            disabled={loading}
          >
            {loading ? <CircularProgress size={22} color="inherit" /> : "Continuar"}
          </Button>
        </Box>
      ) : (
        <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
          <Alert severity="info">
            <AlertTitle>Pagamento seguro</AlertTitle>
            Seus dados de pagamento são processados pela Stripe. Não armazenamos dados de cartão.
          </Alert>
          {error && (
            <Alert severity="error" onClose={() => setError(null)}>
              {error}
            </Alert>
          )}
          <Box id="payment-element" />
          <Button
            variant="contained"
            size="large"
            onClick={handlePay}
            disabled={processing || !stripe}
          >
            {processing ? (
              <CircularProgress size={22} color="inherit" />
            ) : form.is_recurring ? (
              "Ativar dízimo mensal"
            ) : (
              "Pagar dízimo"
            )}
          </Button>
          <Button onClick={() => { setStage("form"); setClientSecret(null); }}>
            Voltar
          </Button>
          <Typography variant="caption" color="text.secondary">
            Aceitamos PIX, cartão de débito e cartão de crédito, conforme disponibilidade.
          </Typography>
        </Box>
      )}
    </Box>
  );
}