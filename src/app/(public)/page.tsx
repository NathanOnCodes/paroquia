import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Grid from "@mui/material/Grid";
import { listPublishedEvents } from "@/features/events/queries";
import { formatDateTime } from "@/lib/format";
import { getSiteSettings } from "@/features/settings/queries";
import EmptyState from "@/design-system/feedback/EmptyState";

export const dynamic = "force-dynamic";

export default async function HomePage() {
  const [events, settings] = await Promise.all([listPublishedEvents(), getSiteSettings()]);

  return (
    <Box>
      <Box sx={{ textAlign: "center", py: { xs: 4, md: 8 } }}>
        <Typography variant="h3" component="h1" sx={{ fontWeight: 700 }}>
          {settings.home_title || `Bem-vindo à ${settings.site_name}`}
        </Typography>
        {settings.city && <Typography variant="subtitle1" color="secondary" sx={{ mt: 1 }}>{settings.city}</Typography>}
        <Typography variant="h6" sx={{ color: "text.secondary", mt: 1, maxWidth: 640, mx: "auto" }}>
          {settings.home_description || settings.description}
        </Typography>
        <Box sx={{ display: "flex", gap: 2, justifyContent: "center", mt: 4, flexWrap: "wrap" }}>
          <Button href="/doacoes" variant="contained" color="secondary" size="large">
            Fazer uma doação
          </Button>
          <Button href="/transparencia" variant="outlined" size="large">
            Ver transparência
          </Button>
        </Box>
      </Box>

      <Box sx={{ mt: 4 }}>
        <Typography variant="h5" sx={{ mb: 2 }}>
          Próximos eventos
        </Typography>
        {events.length === 0 ? (
          <EmptyState title="Nenhum evento publicado" description="A agenda da comunidade será atualizada em breve." />
        ) : (
          <Grid container spacing={3}>
            {events.slice(0, 6).map((event) => (
              <Grid size={{ xs: 12, sm: 6, md: 4 }} key={event.id}>
                <Card sx={{ height: "100%" }}>
                  <CardContent>
                    <Typography variant="subtitle2" color="primary">
                      {formatDateTime(event.starts_at)}
                    </Typography>
                    <Typography variant="h6" sx={{ mt: 1 }}>
                      {event.title}
                    </Typography>
                    {event.location && (
                      <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                        {event.location}
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

      <Box sx={{ mt: 6, display: "flex", justifyContent: "center" }}>
        <Button href="/contato" variant="text">
          Fale conosco
        </Button>
      </Box>
    </Box>
  );
}
