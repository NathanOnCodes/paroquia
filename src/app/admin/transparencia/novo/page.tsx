import Typography from "@mui/material/Typography";
import TextField from "@mui/material/TextField";
import MenuItem from "@mui/material/MenuItem";
import ServerForm from "@/components/admin/ServerForm";
import { createFinancialPeriodAction } from "@/features/transparency/actions";

export default function NewFinancialPeriodPage() {
  return (
    <>
      <Typography variant="h5" sx={{ mb: 2 }}>
        Novo período financeiro
      </Typography>
      <ServerForm action={createFinancialPeriodAction} cancelHref="/admin/transparencia">
        <TextField label="Nome do período" name="name" required fullWidth placeholder="Ex.: Janeiro de 2026" />
        <TextField label="Data inicial" name="start_date" type="date" required fullWidth slotProps={{ inputLabel: { shrink: true } }} />
        <TextField label="Data final" name="end_date" type="date" required fullWidth slotProps={{ inputLabel: { shrink: true } }} />
        <TextField select label="Status" name="status" defaultValue="draft" fullWidth>
          <MenuItem value="draft">Rascunho</MenuItem>
          <MenuItem value="review">Em revisão</MenuItem>
        </TextField>
      </ServerForm>
    </>
  );
}