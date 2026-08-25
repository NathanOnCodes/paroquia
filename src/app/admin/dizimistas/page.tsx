import Link from "next/link";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Paper from "@mui/material/Paper";
import TextField from "@mui/material/TextField";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Chip from "@mui/material/Chip";
import { listDonors } from "@/features/donors/queries";
import PageHeader from "@/design-system/layout/PageHeader";
import { formatDate } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function DizimistasPage({
  searchParams,
}: {
  searchParams: Promise<{ q?: string }>;
}) {
  const { q } = await searchParams;
  const donors = await listDonors(q);

  return (
    <Box>
      <PageHeader
        title="Dizimistas"
        action={
          <Button href="/admin/dizimistas/novo" variant="contained">
            Novo dizimista
          </Button>
        }
      />

      <form action="/admin/dizimistas" method="get" style={{ marginBottom: 16 }}>
        <TextField
          name="q"
          defaultValue={q ?? ""}
          placeholder="Buscar por nome, e-mail ou telefone..."
          size="small"
          sx={{ minWidth: 300 }}
        />
        <Button type="submit" variant="outlined" sx={{ ml: 1 }}>
          Buscar
        </Button>
      </form>

      <TableContainer component={Paper}>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Nome</TableCell>
              <TableCell>E-mail</TableCell>
              <TableCell>Telefone</TableCell>
              <TableCell>Comunidade</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Cadastro</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {donors.length === 0 && (
              <TableRow>
                <TableCell colSpan={6}>Nenhum dizimista encontrado.</TableCell>
              </TableRow>
            )}
{donors.map((donor) => (
              <TableRow key={donor.id} hover sx={{ cursor: "pointer" }}>
                <TableCell>
                  <Link href={`/admin/dizimistas/${donor.id}`} style={{ color: "inherit", textDecoration: "none" }}>
                    <strong>{donor.full_name}</strong>
                  </Link>
                </TableCell>
                <TableCell>{donor.email}</TableCell>
                <TableCell>{donor.phone ?? ""}</TableCell>
                <TableCell>
                  {(donor as unknown as { communities?: { name: string } }).communities?.name ??
                    donor.community_name ??
                    ""}
                </TableCell>
                <TableCell>
                  <Chip
                    size="small"
                    label={donor.is_active ? "Ativo" : "Inativo"}
                    color={donor.is_active ? "success" : "default"}
                  />
                </TableCell>
                <TableCell>{formatDate(donor.created_at)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
}
