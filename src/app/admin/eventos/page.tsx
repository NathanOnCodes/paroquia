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
import { listEventsAdmin } from "@/features/events/queries";
import PageHeader from "@/design-system/layout/PageHeader";
import { formatDateTime } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function AdminEventsPage() {
  const events = await listEventsAdmin();

  return (
    <Box>
      <PageHeader
        title="Eventos"
        action={
          <Button href="/admin/eventos/novo" variant="contained">
            Novo evento
          </Button>
        }
      />
      <TableContainer component={Paper}>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Título</TableCell>
              <TableCell>Data</TableCell>
              <TableCell sx={{ display: { xs: "none", md: "table-cell" } }}>Local</TableCell>
              <TableCell>Status</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {events.length === 0 && (
              <TableRow>
                <TableCell colSpan={4}>Nenhum evento cadastrado.</TableCell>
              </TableRow>
            )}
{events.map((event) => (
              <TableRow key={event.id} hover sx={{ cursor: "pointer" }}>
                <TableCell>
                  <Link href={`/admin/eventos/${event.id}`} style={{ color: "inherit", textDecoration: "none" }}>
                    <strong>{event.title}</strong>
                  </Link>
                </TableCell>
                <TableCell>{formatDateTime(event.starts_at)}</TableCell>
                <TableCell sx={{ display: { xs: "none", md: "table-cell" } }}>{event.location || ""}</TableCell>
                <TableCell>
                  <Chip
                    size="small"
                    label={event.status}
                    color={event.status === "published" ? "success" : event.status === "draft" ? "warning" : "default"}
                  />
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
}
