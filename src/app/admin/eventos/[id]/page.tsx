import { notFound } from "next/navigation";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";
import TextField from "@mui/material/TextField";
import MenuItem from "@mui/material/MenuItem";
import Paper from "@mui/material/Paper";
import ServerForm from "@/components/admin/ServerForm";
import { getEventById } from "@/features/events/queries";
import { updateEventAction } from "@/features/events/actions";
import { formatDateTime } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function EditEventPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const event = await getEventById(id);
  if (!event) notFound();

  const startsLocal = toLocalInput(event.starts_at);

  return (
    <Box>
      <Button href="/admin/eventos" sx={{ mb: 2 }}>
        ← Voltar
      </Button>
      <Typography variant="h5" sx={{ mb: 2 }}>
        Editar evento
      </Typography>

      <Paper sx={{ p: 3, mb: 3 }}>
        <ServerForm action={updateEventAction} successMessage="Evento atualizado!">
          <input type="hidden" name="id" value={event.id} />
          <TextField label="Título" name="title" defaultValue={event.title} required fullWidth />
          <TextField label="Slug" name="slug" defaultValue={event.slug} fullWidth />
          <TextField label="Descrição" name="description" defaultValue={event.description ?? ""} multiline rows={5} fullWidth />
          <TextField label="Data de início" name="starts_at" type="datetime-local" defaultValue={startsLocal} required fullWidth slotProps={{ inputLabel: { shrink: true } }} />
          <TextField label="Data de fim (opcional)" name="ends_at" type="datetime-local" defaultValue={event.ends_at ? toLocalInput(event.ends_at) : ""} fullWidth slotProps={{ inputLabel: { shrink: true } }} />
          <TextField label="Local" name="location" defaultValue={event.location ?? ""} fullWidth />
          <TextField select label="Status" name="status" defaultValue={event.status} fullWidth>
            <MenuItem value="draft">Rascunho</MenuItem>
            <MenuItem value="published">Publicado</MenuItem>
            <MenuItem value="archived">Arquivado</MenuItem>
          </TextField>
          <TextField label="Nova imagem (opcional)" name="image" type="file" fullWidth slotProps={{ inputLabel: { shrink: true } }} />
        </ServerForm>
      </Paper>

      <Typography variant="body2" color="text.secondary">
        Criado em {formatDateTime(event.created_at)}
        {event.published_at && ` · Publicado em ${formatDateTime(event.published_at)}`}
      </Typography>
    </Box>
  );
}

function toLocalInput(iso: string): string {
  const d = new Date(iso);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}