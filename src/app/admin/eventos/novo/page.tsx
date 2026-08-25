import Typography from "@mui/material/Typography";
import TextField from "@mui/material/TextField";
import MenuItem from "@mui/material/MenuItem";
import ServerForm from "@/design-system/forms/ServerForm";
import { createEventAction } from "@/features/events/actions";

export default function NewEventPage() {
  return (
    <>
      <Typography variant="h5" sx={{ mb: 2 }}>
        Novo evento
      </Typography>
      <ServerForm action={createEventAction} cancelHref="/admin/eventos">
        <TextField label="Título" name="title" required fullWidth />
        <TextField label="Slug (opcional)" name="slug" helperText="Deixe em branco para gerar automaticamente." fullWidth />
        <TextField label="Descrição" name="description" multiline rows={5} fullWidth />
        <TextField label="Data de início" name="starts_at" type="datetime-local" required fullWidth slotProps={{ inputLabel: { shrink: true } }} />
        <TextField label="Data de fim (opcional)" name="ends_at" type="datetime-local" fullWidth slotProps={{ inputLabel: { shrink: true } }} />
        <TextField label="Local" name="location" fullWidth />
        <TextField select label="Status" name="status" defaultValue="draft" fullWidth>
          <MenuItem value="draft">Rascunho</MenuItem>
          <MenuItem value="published">Publicado</MenuItem>
        </TextField>
        <TextField label="Imagem (JPG, PNG, WebP)" name="image" type="file" fullWidth slotProps={{ inputLabel: { shrink: true } }} />
      </ServerForm>
    </>
  );
}