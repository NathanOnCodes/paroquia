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
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemText from "@mui/material/ListItemText";
import IconButton from "@mui/material/IconButton";
import PictureAsPdfIcon from "@mui/icons-material/PictureAsPdf";
import ImageIcon from "@mui/icons-material/Image";
import { getPublishedPeriodById } from "@/features/transparency/queries";
import { formatCurrency, formatDate } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function TransparencyDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const data = await getPublishedPeriodById(id);
  if (!data) notFound();

  const { period, entries, documents } = data;

  const expensesByCategory = new Map<string, number>();
  for (const entry of entries) {
    if (entry.type === "expense") {
      expensesByCategory.set(
        entry.category,
        (expensesByCategory.get(entry.category) ?? 0) + entry.amount_cents
      );
    } else if (entry.source === "online_contribution") {
    }
  }

  return (
    <Box sx={{ maxWidth: 860, mx: "auto" }}>
      <Button href="/transparencia" sx={{ mb: 2 }}>
        ← Voltar para transparência
      </Button>
      <Typography variant="h4" sx={{ mb: 0.5 }}>
        {period.name}
      </Typography>
      <Typography color="text.secondary" sx={{ mb: 3 }}>
        {formatDate(period.start_date)} a {formatDate(period.end_date)} · Publicado em{" "}
        {formatDate(period.published_at)}
      </Typography>

      <Box sx={{ display: "flex", gap: 2, flexWrap: "wrap", mb: 4 }}>
        <Paper sx={{ p: 2, flex: 1, minWidth: 180 }}>
          <Typography variant="subtitle2" color="text.secondary">
            Arrecadado
          </Typography>
          <Typography variant="h5" color="success.main">
            {formatCurrency(period.total_income_cents)}
          </Typography>
        </Paper>
        <Paper sx={{ p: 2, flex: 1, minWidth: 180 }}>
          <Typography variant="subtitle2" color="text.secondary">
            Despesas
          </Typography>
          <Typography variant="h5" color="error.main">
            {formatCurrency(period.total_expense_cents)}
          </Typography>
        </Paper>
        <Paper sx={{ p: 2, flex: 1, minWidth: 180 }}>
          <Typography variant="subtitle2" color="text.secondary">
            Saldo
          </Typography>
          <Typography variant="h5" color={period.balance_cents >= 0 ? "primary" : "error"}>
            {formatCurrency(period.balance_cents)}
          </Typography>
        </Paper>
        <Paper sx={{ p: 2, flex: 1, minWidth: 140 }}>
          <Typography variant="subtitle2" color="text.secondary">
            Contribuições
          </Typography>
          <Typography variant="h5">{period.contributions_count}</Typography>
        </Paper>
      </Box>

      <Typography variant="h6" sx={{ mb: 1 }}>
        Lançamentos
      </Typography>
      <TableContainer component={Paper} sx={{ mb: 4 }}>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>Data</TableCell>
              <TableCell>Tipo</TableCell>
              <TableCell>Categoria</TableCell>
              <TableCell>Descrição</TableCell>
              <TableCell align="right">Valor</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {entries.map((entry) => (
              <TableRow key={entry.id}>
                <TableCell>{formatDate(entry.entry_date)}</TableCell>
                <TableCell>
                  <Chip
                    size="small"
                    label={entry.type === "income" ? "Receita" : "Despesa"}
                    color={entry.type === "income" ? "success" : "error"}
                  />
                </TableCell>
                <TableCell>{entry.category}</TableCell>
                <TableCell>{entry.description}</TableCell>
                <TableCell align="right">
                  {entry.type === "income" ? "+" : "−"} {formatCurrency(entry.amount_cents)}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      {expensesByCategory.size > 0 && (
        <Box sx={{ mb: 4 }}>
          <Typography variant="h6" sx={{ mb: 1 }}>
            Despesas por categoria
          </Typography>
          <List>
            {[...expensesByCategory.entries()].map(([category, value]) => (
              <ListItem key={category} divider>
                <ListItemText primary={category} />
                <Typography variant="body1">{formatCurrency(value)}</Typography>
              </ListItem>
            ))}
          </List>
        </Box>
      )}

      <Typography variant="h6" sx={{ mb: 1 }}>
        Documentos DRE
      </Typography>
      {documents.length === 0 ? (
        <Typography color="text.secondary" sx={{ mb: 4 }}>
          Nenhum documento publicado para este período.
        </Typography>
      ) : (
        <List sx={{ mb: 4 }}>
          {documents.map((doc) => (
            <ListItem key={doc.id} secondaryAction={
              <IconButton href={`/api/files/transparency/${doc.id}`} target="_blank" rel="noopener noreferrer" aria-label="Abrir documento">
                {doc.mime_type === "application/pdf" ? <PictureAsPdfIcon /> : <ImageIcon />}
              </IconButton>
            } divider>
              <ListItemText
                primary={doc.original_file_name}
                secondary={`${Math.round(doc.file_size / 1024)} KB`}
              />
            </ListItem>
          ))}
        </List>
      )}
    </Box>
  );
}