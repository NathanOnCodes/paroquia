import Box from "@mui/material/Box";
import Grid from "@mui/material/Grid";
import Paper from "@mui/material/Paper";
import Typography from "@mui/material/Typography";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import { getDashboardSummary } from "@/features/dashboard/queries";
import { formatCurrency } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function DashboardPage() {
  const summary = await getDashboardSummary();

  const cards = [
    { label: "Total arrecadado", value: formatCurrency(summary.total_collected_cents), color: "success.main" },
    { label: "Total de despesas", value: formatCurrency(summary.total_expense_cents), color: "error.main" },
    { label: "Saldo", value: formatCurrency(summary.balance_cents), color: "primary.main" },
    { label: "Dizimistas", value: String(summary.total_donors), color: "text.primary" },
    { label: "Contribuições pagas", value: String(summary.total_paid_contributions), color: "text.primary" },
    { label: "Pagamentos pendentes", value: String(summary.pending_payments), color: "warning.main" },
    { label: "Pagamentos com falha", value: String(summary.failed_payments), color: "error.main" },
    { label: "Recorrências canceladas", value: String(summary.canceled_recurring), color: "text.secondary" },
  ];

  return (
    <Box>
      <Typography variant="h5" sx={{ mb: 3 }}>
        Dashboard
      </Typography>

      <Grid container spacing={2} sx={{ mb: 4 }}>
        {cards.map((card) => (
          <Grid size={{ xs: 12, sm: 6, md: 3 }} key={card.label}>
            <Paper sx={{ p: 2 }}>
              <Typography variant="subtitle2" color="text.secondary">
                {card.label}
              </Typography>
              <Typography variant="h6" sx={{ color: card.color }}>
                {card.value}
              </Typography>
            </Paper>
          </Grid>
        ))}
      </Grid>

      <Grid container spacing={3}>
        <Grid size={{ xs: 12, md: 7 }}>
          <Typography variant="h6" sx={{ mb: 1 }}>
            Comunidades que mais contribuem
          </Typography>
          <TableContainer component={Paper}>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Comunidade</TableCell>
                  <TableCell align="right">Valor</TableCell>
                  <TableCell align="right">Contribuições</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {summary.community_ranking.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={3}>Nenhuma contribuição registrada.</TableCell>
                  </TableRow>
                )}
                {summary.community_ranking.map((row) => (
                  <TableRow key={row.community_id ?? "none"}>
                    <TableCell>{row.community_name}</TableCell>
                    <TableCell align="right">{formatCurrency(row.total_cents)}</TableCell>
                    <TableCell align="right">{row.count}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </Grid>
        <Grid size={{ xs: 12, md: 5 }}>
          <Typography variant="h6" sx={{ mb: 1 }}>
            Evolução mensal
          </Typography>
          <TableContainer component={Paper}>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Mês</TableCell>
                  <TableCell align="right">Total</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {summary.monthly_evolution.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={2}>Sem dados.</TableCell>
                  </TableRow>
                )}
                {summary.monthly_evolution.map((row) => (
                  <TableRow key={row.month}>
                    <TableCell>{formatMonth(row.month)}</TableCell>
                    <TableCell align="right">{formatCurrency(row.total_cents)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </Grid>
      </Grid>
    </Box>
  );
}

function formatMonth(month: string): string {
  const [year, m] = month.split("-");
  const names = [
    "Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho",
    "Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro",
  ];
  return `${names[Number(m) - 1] ?? m} ${year}`;
}