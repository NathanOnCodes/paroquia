import Link from "next/link";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Paper from "@mui/material/Paper";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Chip from "@mui/material/Chip";
import { listFinancialPeriodsAdmin } from "@/features/transparency/queries.admin";
import PageHeader from "@/design-system/layout/PageHeader";
import { changeFinancialPeriodStatusAction } from "@/features/transparency/actions";
import { voidAction } from "@/lib/actions";
import { formatDate } from "@/lib/format";
import { requireRole } from "@/lib/auth/authorization";

export const dynamic = "force-dynamic";

export default async function AdminTransparencyPage() {
  const user = await requireRole("assistente");
  const periods = await listFinancialPeriodsAdmin();

  return (
    <Box>
      <PageHeader
        title="Transparência financeira"
        action={
          <Button href="/admin/transparencia/novo" variant="contained">
            Novo período
          </Button>
        }
      />
      <TableContainer component={Paper}>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Nome</TableCell>
              <TableCell>Período</TableCell>
              <TableCell>Status</TableCell>
              <TableCell sx={{ width: 320 }}>Ações</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {periods.length === 0 && (
              <TableRow>
                <TableCell colSpan={4}>Nenhum período cadastrado.</TableCell>
              </TableRow>
            )}
            {periods.map((period) => (
              <TableRow key={period.id}>
                <TableCell>
                  <Link href={`/admin/transparencia/${period.id}`} style={{ color: "inherit" }}>
                    <strong>{period.name}</strong>
                  </Link>
                </TableCell>
                <TableCell>
                  {formatDate(period.start_date)} a {formatDate(period.end_date)}
                </TableCell>
                <TableCell>
                  <Chip
                    size="small"
                    label={period.status}
                    color={period.status === "published" ? "success" : period.status === "review" ? "warning" : "default"}
                  />
                </TableCell>
                <TableCell>
                  <Box sx={{ display: "flex", gap: 1, flexWrap: "wrap" }}>
                    <Button size="small" href={`/admin/transparencia/${period.id}`} variant="outlined">
                      Abrir
                    </Button>
                    {period.status === "draft" && (
                      <form action={voidAction(changeFinancialPeriodStatusAction)}>
                        <input type="hidden" name="id" value={period.id} />
                        <input type="hidden" name="status" value="review" />
                        <Button size="small" type="submit" variant="outlined" color="warning">
                          Enviar para revisão
                        </Button>
                      </form>
                    )}
                    {(period.status === "review" || period.status === "draft") &&
                      user.profile.role === "admin" && (
                        <form action={voidAction(changeFinancialPeriodStatusAction)}>
                          <input type="hidden" name="id" value={period.id} />
                          <input type="hidden" name="status" value="published" />
                          <Button size="small" type="submit" variant="contained" color="success">
                            Publicar
                          </Button>
                        </form>
                      )}
                    {period.status === "published" && user.profile.role === "admin" && (
                      <form action={voidAction(changeFinancialPeriodStatusAction)}>
                        <input type="hidden" name="id" value={period.id} />
                        <input type="hidden" name="status" value="archived" />
                        <Button size="small" type="submit" variant="outlined" color="error">
                          Arquivar
                        </Button>
                      </form>
                    )}
                  </Box>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
}