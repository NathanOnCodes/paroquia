import { notFound } from "next/navigation";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";
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
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemText from "@mui/material/ListItemText";
import ServerForm from "@/components/admin/ServerForm";
import { getFinancialPeriodById } from "@/features/auth/queries.admin";
import { listDocumentsForPeriod } from "@/features/transparency/queries.admin";
import {
  addFinancialEntryAction,
  deleteFinancialEntryAction,
  changeFinancialPeriodStatusAction,
  uploadTransparencyDocumentAction,
} from "@/features/transparency/actions";
import { requireRole } from "@/lib/auth/authorization";
import { voidAction } from "@/lib/actions";
import { formatCurrency, formatDate } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function FinancialPeriodDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const user = await requireRole("assistente");
  const data = await getFinancialPeriodById(id);
  if (!data) notFound();

  const documents = await listDocumentsForPeriod(id);
  const { period, entries } = data;

  const totalIncome = entries.filter((e) => e.type === "income").reduce((a, e) => a + e.amount_cents, 0);
  const totalExpense = entries.filter((e) => e.type === "expense").reduce((a, e) => a + e.amount_cents, 0);

  return (
    <Box>
      <Button href="/admin/transparencia" sx={{ mb: 2 }}>
        ← Voltar
      </Button>
      <Typography variant="h5" sx={{ mb: 0.5 }}>
        {period.name}
      </Typography>
      <Typography color="text.secondary" sx={{ mb: 2 }}>
        {formatDate(period.start_date)} a {formatDate(period.end_date)} · Status:{" "}
        <Chip size="small" label={period.status} color={period.status === "published" ? "success" : period.status === "review" ? "warning" : "default"} />
      </Typography>

      <Box sx={{ display: "flex", gap: 2, flexWrap: "wrap", mb: 3 }}>
        <Paper sx={{ p: 2 }}>
          <Typography variant="subtitle2" color="text.secondary">Receitas</Typography>
          <Typography variant="h6" color="success.main">{formatCurrency(totalIncome)}</Typography>
        </Paper>
        <Paper sx={{ p: 2 }}>
          <Typography variant="subtitle2" color="text.secondary">Despesas</Typography>
          <Typography variant="h6" color="error.main">{formatCurrency(totalExpense)}</Typography>
        </Paper>
        <Paper sx={{ p: 2 }}>
          <Typography variant="subtitle2" color="text.secondary">Saldo</Typography>
          <Typography variant="h6">{formatCurrency(totalIncome - totalExpense)}</Typography>
        </Paper>
      </Box>

      <Box sx={{ display: "flex", gap: 1, flexWrap: "wrap", mb: 4 }}>
        {period.status === "draft" && (
          <form action={voidAction(changeFinancialPeriodStatusAction)}>
            <input type="hidden" name="id" value={period.id} />
            <input type="hidden" name="status" value="review" />
            <Button type="submit" variant="outlined" color="warning">
              Enviar para revisão
            </Button>
          </form>
        )}
        {period.status !== "published" && user.profile.role === "admin" && (
          <form action={voidAction(changeFinancialPeriodStatusAction)}>
            <input type="hidden" name="id" value={period.id} />
            <input type="hidden" name="status" value="published" />
            <Button type="submit" variant="contained" color="success">
              Publicar
            </Button>
          </form>
        )}
        {period.status === "published" && user.profile.role === "admin" && (
          <form action={voidAction(changeFinancialPeriodStatusAction)}>
            <input type="hidden" name="id" value={period.id} />
            <input type="hidden" name="status" value="archived" />
            <Button type="submit" variant="outlined" color="error">
              Arquivar
            </Button>
          </form>
        )}
      </Box>

      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" sx={{ mb: 2 }}>
          Adicionar lançamento
        </Typography>
        <ServerForm action={addFinancialEntryAction} successMessage="Lançamento adicionado!">
          <input type="hidden" name="financial_period_id" value={period.id} />
          <TextField select label="Tipo" name="type" defaultValue="income" fullWidth>
            <MenuItem value="income">Receita</MenuItem>
            <MenuItem value="expense">Despesa</MenuItem>
          </TextField>
          <TextField label="Categoria" name="category" required fullWidth placeholder="Ex.: Dízimos, Manutenção" />
          <TextField label="Descrição" name="description" required fullWidth />
          <TextField label="Valor (em centavos ou reais?)" name="amount_cents" required fullWidth placeholder="Valor em centavos, ex.: 5000 = R$ 50,00" />
          <TextField label="Data do lançamento" name="entry_date" type="date" required fullWidth slotProps={{ inputLabel: { shrink: true } }} />
        </ServerForm>
      </Paper>

      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" sx={{ mb: 2 }}>
          Upload de documento DRE
        </Typography>
        <ServerForm action={uploadTransparencyDocumentAction} successMessage="Documento enviado!" submitLabel="Enviar arquivo">
          <input type="hidden" name="period_id" value={period.id} />
          <TextField label="Arquivo (PDF, JPG, PNG, WebP)" name="file" type="file" fullWidth slotProps={{ inputLabel: { shrink: true } }} />
        </ServerForm>
      </Paper>

      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" sx={{ mb: 2 }}>
          Documentos
        </Typography>
        {documents.length === 0 ? (
          <Typography color="text.secondary">Nenhum documento enviado.</Typography>
        ) : (
          <List>
            {documents.map((doc) => (
              <ListItem key={doc.id} divider>
                <ListItemText
                  primary={doc.original_file_name}
                  secondary={`${doc.mime_type} · ${Math.round(doc.file_size / 1024)} KB`}
                />
              </ListItem>
            ))}
          </List>
        )}
      </Paper>

      <Paper sx={{ p: 3 }}>
        <Typography variant="h6" sx={{ mb: 2 }}>
          Lançamentos
        </Typography>
        <TableContainer>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Data</TableCell>
                <TableCell>Tipo</TableCell>
                <TableCell>Categoria</TableCell>
                <TableCell>Descrição</TableCell>
                <TableCell align="right">Valor</TableCell>
                <TableCell></TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {entries.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6}>Nenhum lançamento.</TableCell>
                </TableRow>
              )}
              {entries.map((entry) => (
                <TableRow key={entry.id}>
                  <TableCell>{formatDate(entry.entry_date)}</TableCell>
                  <TableCell>
                    <Chip size="small" label={entry.type === "income" ? "Receita" : "Despesa"} color={entry.type === "income" ? "success" : "error"} />
                  </TableCell>
                  <TableCell>{entry.category}</TableCell>
                  <TableCell>{entry.description}</TableCell>
                  <TableCell align="right">{formatCurrency(entry.amount_cents)}</TableCell>
                  <TableCell>
                    <form action={voidAction(deleteFinancialEntryAction)}>
                      <input type="hidden" name="id" value={entry.id} />
                      <input type="hidden" name="period_id" value={period.id} />
                      <Button size="small" color="error" type="submit">
                        Excluir
                      </Button>
                    </form>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </Paper>
    </Box>
  );
}