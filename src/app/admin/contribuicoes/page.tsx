import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Paper from "@mui/material/Paper";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Chip from "@mui/material/Chip";
import { listContributionsAdmin } from "@/features/contributions/queries";
import PageHeader from "@/design-system/layout/PageHeader";
import { formatCurrency, formatDateTime } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function ContribuicoesPage() {
  const contributions = await listContributionsAdmin();

  return (
    <Box>
      <PageHeader title="Contribuições" />
      <TableContainer component={Paper}>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Dizimista</TableCell>
              <TableCell>Comunidade</TableCell>
              <TableCell>Valor</TableCell>
              <TableCell>Tipo</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Data</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {contributions.length === 0 && (
              <TableRow>
                <TableCell colSpan={6}>Nenhuma contribuição registrada.</TableCell>
              </TableRow>
            )}
            {contributions.map((c) => (
              <TableRow key={c.id}>
                <TableCell>
                  {c.donors?.full_name ?? ""}
                  <Typography variant="caption" sx={{ display: "block" }} color="text.secondary">
                    {c.donors?.email}
                  </Typography>
                </TableCell>
                <TableCell>
                  {c.communities?.name ?? c.community_name ?? c.donors?.community_name ?? ""}
                </TableCell>
                <TableCell>{formatCurrency(c.amount_cents)}</TableCell>
                <TableCell>{c.payment_type === "recurring" ? "Recorrente" : "Avulso"}</TableCell>
                <TableCell>
                  <Chip
                    size="small"
                    label={c.status}
                    color={c.status === "paid" ? "success" : c.status === "pending" ? "warning" : c.status === "failed" || c.status === "canceled" ? "error" : "default"}
                  />
                </TableCell>
                <TableCell>{formatDateTime(c.created_at)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
}
