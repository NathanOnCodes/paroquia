import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Button from "@mui/material/Button";
import Grid from "@mui/material/Grid";
import { listPublishedEvents } from "@/features/events/queries";
import { formatDateTime } from "@/lib/format";
import { getSiteSettings } from "@/features/settings/queries";
import PageHeader from "@/design-system/layout/PageHeader";
import EmptyState from "@/design-system/feedback/EmptyState";

export const dynamic = "force-dynamic";

export default async function EventsPage() {
  const [events, settings] = await Promise.all([listPublishedEvents(), getSiteSettings()]);

  return (
    <Box>
      <PageHeader title={`Eventos da ${settings.site_name}`} description="Confira os próximos encontros e atividades da comunidade." />
      {events.length === 0 ? (
        <EmptyState title="Nenhum evento publicado" description="A agenda da comunidade será atualizada em breve." />
      ) : (
        <Grid container spacing={3}>
          {events.map((event) => (
            <Grid size={{ xs: 12, sm: 6, md: 4 }} key={event.id}>
              <Card sx={{ height: "100%" }}>
                <CardContent>
                  <Typography variant="subtitle2" color="primary">
                    {formatDateTime(event.starts_at)}
                  </Typography>
                  <Typography variant="h6" sx={{ mt: 1 }}>
                    {event.title}
                  </Typography>
                  {event.description && (
                    <Typography
                      variant="body2"
                      color="text.secondary"
                      sx={{ mt: 1, overflow: "hidden", textOverflow: "ellipsis", display: "-webkit-box", WebkitLineClamp: 3, WebkitBoxOrient: "vertical" }}
                    >
                      {event.description}
                    </Typography>
                  )}
                  {event.location && (
                    <Typography variant="body2" sx={{ mt: 1 }}>
                      📍 {event.location}
                    </Typography>
                  )}
                  <Button href={`/eventos/${event.slug}`} sx={{ mt: 2 }}>
                    Ver detalhes
                  </Button>
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>
      )}
    </Box>
  );
}
