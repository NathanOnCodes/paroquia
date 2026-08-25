import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Button from "@mui/material/Button";
import Grid from "@mui/material/Grid";
import PageHeader from "@/design-system/layout/PageHeader";
import EmptyState from "@/design-system/feedback/EmptyState";
import { listPublishedPeriods } from "@/features/transparency/queries";
import { formatCurrency, formatDate } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function TransparencyPage() {
  const periods = await listPublishedPeriods();

  return (
    <Box>
      <PageHeader
        title="Transparência financeira"
        description="Acompanhe quanto a paróquia arrecada, quanto gasta e confira os documentos de prestação de contas (DRE). Nenhum dado pessoal é divulgado."
      />

      {periods.length === 0 ? (
        <EmptyState
          title="Nada publicado ainda"
          description="Os períodos de prestação de contas aparecerão aqui assim que forem publicados."
        />
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