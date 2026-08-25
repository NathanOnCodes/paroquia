import { notFound } from "next/navigation";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";
import TextField from "@mui/material/TextField";
import MenuItem from "@mui/material/MenuItem";
import Paper from "@mui/material/Paper";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Chip from "@mui/material/Chip";
import Divider from "@mui/material/Divider";
import ServerForm from "@/design-system/forms/ServerForm";
import { getDonorById } from "@/features/donors/queries";
import { listAllCommunities } from "@/features/communities/queries.admin";
import { updateDonorAction, setDonorStatusAction } from "@/features/donors/actions";
import { listContributionsByDonor } from "@/features/contributions/queries";
import { listRecurringForDonor } from "@/features/payments/services";
import { formatCurrency, formatDateTime } from "@/lib/format";
import { voidAction } from "@/lib/actions";

export const dynamic = "force-dynamic";

export default async function DonorDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const donor = await getDonorById(id);
  if (!donor) notFound();

  const communities = await listAllCommunities();
  const contributions = await listContributionsByDonor(id);
  const recurring = await listRecurringForDonor(id);

  return (
    <Box>
      <Button href="/admin/dizimistas" sx={{ mb: 2 }}>
        ← Voltar
      </Button>
      <Typography variant="h5" sx={{ mb: 3 }}>
        {donor.full_name}
      </Typography>

      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" sx={{ mb: 2 }}>
          Dados cadastrais
        </Typography>
        <ServerForm action={updateDonorAction} successMessage="Dados atualizados!">
          <input type="hidden" name="id" value={donor.id} />
          <TextField label="Nome completo" name="full_name" defaultValue={donor.full_name} required fullWidth />
          <TextField label="E-mail" name="email" type="email" defaultValue={donor.email} required fullWidth />
          <TextField label="Telefone" name="phone" defaultValue={donor.phone ?? ""} fullWidth />
          <TextField select label="Comunidade" name="community_id" defaultValue={donor.community_id ?? ""} fullWidth>
            <MenuItem value="">Sem comunidade</MenuItem>
            {communities.map((c) => (
              <MenuItem key={c.id} value={c.id}>
                {c.name}
              </MenuItem>
            ))}
          </TextField>
          <TextField
            label="Comunidade fora da lista (opcional)"
            name="community_name"
            defaultValue={donor.community_name ?? ""}
            helperText="Use quando o dizimista pertence a outra paróquia ou cidade."
            fullWidth
          />
          <TextField label="Cidade" name="city" defaultValue={donor.city ?? ""} fullWidth />
          <TextField label="Bairro" name="neighborhood" defaultValue={donor.neighborhood ?? ""} fullWidth />
        </ServerForm>
      </Paper>

      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" sx={{ mb: 2 }}>
          Situação do cadastro
        </Typography>
        <form action={voidAction(setDonorStatusAction)}>
          <input type="hidden" name="id" value={donor.id} />
          <input type="hidden" name="is_active" value={donor.is_active ? "false" : "true"} />
          <Button type="submit" color={donor.is_active ? "error" : "success"} variant="outlined">
            {donor.is_active ? "Inativar cadastro" : "Reativar cadastro"}
          </Button>
        </form>
      </Paper>

      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" sx={{ mb: 2 }}>
          Recorrências
        </Typography>
        {recurring.length === 0 ? (
          <Typography color="text.secondary">Nenhuma recorrência ativa.</Typography>
        ) : (
          <TableContainer>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Valor</TableCell>
                  <TableCell>Status</TableCell>
                  <TableCell>Próxima cobrança</TableCell>
                  <TableCell>Cancelamento solicitado</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {recurring.map((sub) => (
                  <TableRow key={sub.id}>
                    <TableCell>{formatCurrency(sub.amount_cents)}</TableCell>
                    <TableCell>
                      <Chip size="small" label={sub.status} color={sub.status === "active" ? "success" : sub.status === "canceled" ? "default" : "warning"} />
                    </TableCell>
                    <TableCell>{formatDateTime(sub.current_period_end)}</TableCell>
                    <TableCell>{formatDateTime(sub.cancellation_requested_at)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </Paper>

      <Paper sx={{ p: 3 }}>
        <Typography variant="h6" sx={{ mb: 2 }}>
          Histórico de contribuições
        </Typography>
        {contributions.length === 0 ? (
          <Typography color="text.secondary">Nenhuma contribuição registrada.</Typography>
        ) : (
          <TableContainer>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Data</TableCell>
                  <TableCell>Valor</TableCell>
                  <TableCell>Tipo</TableCell>
                  <TableCell>Status</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {contributions.map((c) => (
                  <TableRow key={c.id}>
                    <TableCell>{formatDateTime(c.created_at)}</TableCell>
                    <TableCell>{formatCurrency(c.amount_cents)}</TableCell>
                    <TableCell>{c.payment_type === "recurring" ? "Recorrente" : "Avulso"}</TableCell>
                    <TableCell>
                      <Chip size="small" label={c.status} color={c.status === "paid" ? "success" : c.status === "pending" ? "warning" : "default"} />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </Paper>
      <Divider sx={{ mt: 4 }} />
    </Box>
  );
}