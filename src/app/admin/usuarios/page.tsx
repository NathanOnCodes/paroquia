import { redirect } from "next/navigation";
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
import MenuItem from "@mui/material/MenuItem";
import ServerForm from "@/components/admin/ServerForm";
import { listStaffUsers } from "@/features/auth/queries.admin";
import { createStaffUserAction, updateStaffUserAction } from "@/features/auth/actions";
import { requireRole } from "@/lib/auth/authorization";
import { ROLE_LABELS } from "@/lib/auth/roles";
import { formatDate } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function AdminUsersPage() {
  const actor = await requireRole("admin");
  if (actor.profile.role !== "admin") redirect("/admin");

  const users = await listStaffUsers();

  return (
    <Box>
      <Typography variant="h5" sx={{ mb: 2 }}>
        Usuários administrativos
      </Typography>

      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" sx={{ mb: 2 }}>
          Novo usuário
        </Typography>
        <ServerForm action={createStaffUserAction} submitLabel="Criar usuário" successMessage="Usuário criado!">
          <TextField label="Nome completo" name="full_name" required fullWidth />
          <TextField label="E-mail" name="email" type="email" required fullWidth />
          <TextField label="Senha inicial" name="password" type="password" required fullWidth helperText="Mínimo de 8 caracteres." />
          <TextField select label="Papel" name="role" defaultValue="assistente" fullWidth>
            <MenuItem value="assistente">Assistente</MenuItem>
            <MenuItem value="admin">Pároco</MenuItem>
          </TextField>
        </ServerForm>
      </Paper>

      <TableContainer component={Paper}>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Nome</TableCell>
              <TableCell>E-mail</TableCell>
              <TableCell>Papel</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Alterar papel/status</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {users.map((user) => (
              <TableRow key={user.id}>
                <TableCell>{user.full_name}</TableCell>
                <TableCell>{user.email}</TableCell>
                <TableCell>
                  <Chip size="small" label={ROLE_LABELS[user.role]} color={user.role === "admin" ? "primary" : "default"} />
                </TableCell>
                <TableCell>
                  <Chip size="small" label={user.is_active ? "Ativo" : "Inativo"} color={user.is_active ? "success" : "default"} />
                </TableCell>
                <TableCell>
                  {user.id === actor.id ? (
                    <Typography variant="body2" color="text.secondary">
                      (você) · criado em {formatDate(user.created_at)}
                    </Typography>
                  ) : (
                    <ServerForm
                      action={updateStaffUserAction}
                      submitLabel="Salvar"
                      successMessage="Usuário atualizado!"
                    >
                      <input type="hidden" name="id" value={user.id} />
                      <Box sx={{ display: "flex", gap: 1 }}>
                        <TextField size="small" label="Nome" name="full_name" defaultValue={user.full_name} />
                        <TextField select size="small" label="Papel" name="role" defaultValue={user.role} sx={{ minWidth: 130 }}>
                          <MenuItem value="assistente">Assistente</MenuItem>
                          <MenuItem value="admin">Pároco</MenuItem>
                        </TextField>
                        <TextField select size="small" label="Status" name="is_active" defaultValue={user.is_active ? "on" : "off"} sx={{ minWidth: 100 }}>
                          <MenuItem value="on">Ativo</MenuItem>
                          <MenuItem value="off">Inativo</MenuItem>
                        </TextField>
                      </Box>
                    </ServerForm>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
}