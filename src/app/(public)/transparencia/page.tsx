import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Button from "@mui/material/Button";
import Grid from "@mui/material/Grid";
import { listPublishedPeriods } from "@/features/transparency/queries";
import { formatCurrency, formatDate } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function TransparencyPage() {
  const periods = await listPublishedPeriods();

  return (
    <Box>
      <Typography variant="h4" sx={{ mb: 1 }}>
        Transparência financeira
      </Typography>
      <Typography color="text.secondary" sx={{ mb: 3 }}>
        Acompanhe quanto a paróquia arrecada, quanto gasta e confira os
        documentos de prestação de contas (DRE). Nenhum dado pessoal é divulgado.
      </Typography>

      {periods.length === 0 ? (
        <AlertBox />
      ) : (
        <Grid container spacing={3}>
          {periods.map((period) => (
            <Grid size={{ xs: 12, sm: 6, md: 4 }} key={period.id}>
              <Card sx={{ height: "100%" }}>
                <CardContent>
                  <Typography variant="h6">{period.name}</Typography>
                  <Typography variant="body2" color="text.secondary">
                    {formatDate(period.start_date)} a {formatDate(period.end_date)}
                  </Typography>
                  <Box sx={{ mt: 2, display: "grid", gap: 0.5 }}>
                    <Typography variant="body2">
                      Arrecadado:{" "}
                      <strong>{formatCurrency(period.total_income_cents)}</strong>
                    </Typography>
                    <Typography variant="body2">
                      Despesas:{" "}
                      <strong>{formatCurrency(period.total_expense_cents)}</strong>
                    </Typography>
                    <Typography variant="body2">
                      Saldo:{" "}
                      <Box
                        component="strong"
                        sx={{ color: period.balance_cents >= 0 ? "success.main" : "error.main" }}
                      >
                        {formatCurrency(period.balance_cents)}
                      </Box>
                    </Typography>
                  </Box>
                  <Button href={`/transparencia/${period.id}`} sx={{ mt: 2 }}>
                    Ver detalhes
                  </Button>
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>
      )}
    </Box>
  );
}

function AlertBox() {
  return (
    <Typography color="text.secondary">
      Nenhum período de prestação de contas publicado no momento.
    </Typography>
  );
}