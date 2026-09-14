import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Button from "@mui/material/Button";
import Alert from "@mui/material/Alert";
import Stack from "@mui/material/Stack";
import Divider from "@mui/material/Divider";
import PageHeader from "@/design-system/layout/PageHeader";
import ActionForm from "@/design-system/forms/ActionForm";
import RecurringRequestForm from "@/features/recurring/components/RecurringRequestForm";
import { getRecurringSession, cancelRecurringAction } from "@/features/recurring/actions";
import { formatCurrency, formatDateTime } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function RecurringManagePage({
  searchParams,
}: {
  searchParams: Promise<{ error?: string; success?: string }>;
}) {
  const { error, success } = await searchParams;
  const session = await getRecurringSession();

  return (
    <Box sx={{ maxWidth: 640, mx: "auto" }}>
      <PageHeader
        title="Gerenciar minha doação recorrente"
        description="Acesse suas recorrências e cancele quando desejar, sem precisar de senha."
      />

      {error === "invalid" && (
        <Alert severity="error" sx={{ mb: 2 }}>
          Link inválido ou expirado. Solicite um novo link de acesso.
        </Alert>
      )}
      {success && (
        <Alert severity="success" sx={{ mb: 2 }}>
          Recorrência cancelada com sucesso.
        </Alert>
      )}

      {!session.donorId ? (
        <RecurringRequestForm />
      ) : (
        <Stack spacing={2}>
          {session.subscriptions.length === 0 && (
            <Alert severity="info">
              Você não possui doações recorrentes ativas no momento.
            </Alert>
          )}
          {session.subscriptions
            .filter((sub) => sub.status !== "canceled")
            .map((sub) => (
              <Card key={sub.id}>
                <CardContent>
                  <Typography variant="h6" sx={{ mb: 1 }}>
                    Doação mensal · {formatCurrency(sub.amount_cents)}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    Status: {statusLabel(sub.status)}
                  </Typography>
                  {sub.current_period_end && (
                    <Typography variant="body2" color="text.secondary">
                      Próxima cobrança prevista: {formatDateTime(sub.current_period_end)}
                    </Typography>
                  )}
                  {sub.cancellation_requested_at && (
                    <Typography variant="body2" color="warning.main" sx={{ mt: 0.5 }}>
                      Cancelamento solicitado em {formatDateTime(sub.cancellation_requested_at)}.
                      A recorrência será encerrada ao final do ciclo atual.
                    </Typography>
                  )}
                  <Divider sx={{ my: 1.5 }} />
                  {sub.status === "cancel_at_period_end" ? (
                    <Alert severity="info">
                      Cancelamento em andamento para o fim do ciclo atual.
                    </Alert>
                  ) : (
                    <ActionForm action={cancelRecurringAction}>
                      <input type="hidden" name="subscription_id" value={sub.stripe_subscription_id} />
                      <Button type="submit" color="error" variant="outlined">
                        Cancelar recorrência
                      </Button>
                    </ActionForm>
                  )}
                </CardContent>
              </Card>
            ))}
        </Stack>
      )}
    </Box>
  );
}

function statusLabel(status: string): string {
  switch (status) {
    case "active":
      return "Ativa";
    case "past_due":
      return "Pagamento pendente";
    case "cancel_at_period_end":
      return "Cancelamento ao fim do ciclo";
    case "canceled":
      return "Cancelada";
    case "incomplete":
      return "Em ativação";
    default:
      return status;
  }
}
