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
import TextField from "@mui/material/TextField";
import Button from "@mui/material/Button";
import { voidAction } from "@/lib/actions";
import ServerForm from "@/design-system/forms/ServerForm";
import { listCommunitiesAdmin } from "@/features/communities/queries.admin";
import PageHeader from "@/design-system/layout/PageHeader";
import {
  createCommunityAction,
  setCommunityStatusAction,
} from "@/features/communities/actions";

export const dynamic = "force-dynamic";

export default async function ComunidadesPage() {
  const communities = await listCommunitiesAdmin();

  return (
    <Box>
      <PageHeader title="Comunidades" />

      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" sx={{ mb: 2 }}>
          Nova comunidade
        </Typography>
        <ServerForm action={createCommunityAction} submitLabel="Criar">
          <TextField label="Nome" name="name" required fullWidth />
          <TextField label="Descrição" name="description" multiline rows={2} fullWidth />
          <TextField label="Cidade" name="city" fullWidth />
        </ServerForm>
      </Paper>

      <TableContainer component={Paper}>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Nome</TableCell>
              <TableCell sx={{ display: { xs: "none", sm: "table-cell" } }}>Status</TableCell>
              <TableCell sx={{ minWidth: 120 }}>Ações</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {communities.map((community) => (
              <TableRow key={community.id}>
                <TableCell>
                  <strong>{community.name}</strong>
                  {community.description && (
                    <Typography variant="body2" color="text.secondary">
                      {community.description}
                    </Typography>
                  )}
                  {community.city && (
                    <Typography variant="caption" color="text.secondary">
                      {community.city}
                    </Typography>
                  )}
                </TableCell>
                <TableCell sx={{ display: { xs: "none", sm: "table-cell" } }}>
                  <Chip size="small" label={community.is_active ? "Ativa" : "Inativa"} color={community.is_active ? "success" : "default"} />
                </TableCell>
                <TableCell>
                  <Box sx={{ display: "flex", gap: 1, flexDirection: "column" }}>
                    <Button size="small" component="a" href={`/admin/comunidades/${community.id}`} variant="outlined">
                      Editar
                    </Button>
                    <form action={voidAction(setCommunityStatusAction)}>
                      <input type="hidden" name="id" value={community.id} />
                      <input type="hidden" name="is_active" value={community.is_active ? "false" : "true"} />
                      <Button size="small" color={community.is_active ? "error" : "success"} variant="outlined">
                        {community.is_active ? "Inativar" : "Ativar"}
                      </Button>
                    </form>
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