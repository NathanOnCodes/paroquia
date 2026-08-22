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
import Button from "@mui/material/Button";
import { listContactsAdmin } from "@/features/contact/queries.admin";
import { setContactStatusAction } from "@/features/contact/actions";
import { voidAction } from "@/lib/actions";
import { formatDateTime } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function AdminContactsPage() {
  const contacts = await listContactsAdmin();

  return (
    <Box>
      <Typography variant="h5" sx={{ mb: 2 }}>
        Mensagens de contato
      </Typography>
      <TableContainer component={Paper}>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Nome</TableCell>
              <TableCell>Contato</TableCell>
              <TableCell>Mensagem</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Data</TableCell>
              <TableCell sx={{ width: 180 }}>Ações</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {contacts.length === 0 && (
              <TableRow>
                <TableCell colSpan={6}>Nenhuma mensagem recebida.</TableCell>
              </TableRow>
            )}
            {contacts.map((contact) => (
              <TableRow key={contact.id}>
                <TableCell>{contact.name}</TableCell>
                <TableCell>
                  {contact.email}
                  {contact.phone && (
                    <Typography variant="caption" sx={{ display: "block" }} color="text.secondary">
                      {contact.phone}
                    </Typography>
                  )}
                </TableCell>
                <TableCell sx={{ maxWidth: 320 }}>
                  <Typography variant="body2" sx={{ whiteSpace: "pre-wrap" }}>
                    {contact.message}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Chip
                    size="small"
                    label={contact.status}
                    color={contact.status === "new" ? "warning" : contact.status === "resolved" ? "success" : "default"}
                  />
                </TableCell>
                <TableCell>{formatDateTime(contact.created_at)}</TableCell>
                <TableCell>
                  <Box sx={{ display: "flex", gap: 0.5, flexWrap: "wrap" }}>
                    {contact.status === "new" && (
                      <form action={voidAction(setContactStatusAction)}>
                        <input type="hidden" name="id" value={contact.id} />
                        <input type="hidden" name="status" value="in_progress" />
                        <Button size="small" type="submit" variant="outlined">
                          Em atendimento
                        </Button>
                      </form>
                    )}
                    {contact.status !== "resolved" && contact.status !== "archived" && (
                      <form action={voidAction(setContactStatusAction)}>
                        <input type="hidden" name="id" value={contact.id} />
                        <input type="hidden" name="status" value="resolved" />
                        <Button size="small" type="submit" variant="outlined" color="success">
                          Resolver
                        </Button>
                      </form>
                    )}
                    {contact.status !== "archived" && (
                      <form action={voidAction(setContactStatusAction)}>
                        <input type="hidden" name="id" value={contact.id} />
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