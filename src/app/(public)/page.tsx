import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import Grid from "@mui/material/Grid";
import { listPublishedEvents } from "@/features/events/queries";
import { formatDateTime } from "@/lib/format";

export const dynamic = "force-dynamic";

export default async function HomePage() {
  const events = await listPublishedEvents();

  return (
    <Box>
      <Box sx={{ textAlign: "center", py: { xs: 4, md: 8 } }}>
        <Typography variant="h3" component="h1" sx={{ fontWeight: 700 }}>
          Bem-vindo à Paróquia São Benedito e Menino Jesus
        </Typography>
        <Typography variant="subtitle1" color="primary" sx={{ mt: 1 }}>
          Francisco Morato - SP
        </Typography>
        <Typography variant="h6" sx={{ color: "text.secondary", mt: 1, maxWidth: 640, mx: "auto" }}>
          Um espaço de fé, comunidade e transparência. Participe dos eventos,
          contribua com o dízimo online e acompanhe a prestação de contas.
        </Typography>
        <Box sx={{ display: "flex", gap: 2, justifyContent: "center", mt: 4, flexWrap: "wrap" }}>
          <Button href="/dizimo" variant="contained" size="large">
            Fazer meu dízimo
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
          <Typography color="text.secondary">Nenhum evento publicado no momento.</Typography>
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
