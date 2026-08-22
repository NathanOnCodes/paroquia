import { notFound } from "next/navigation";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";
import { getPublishedEventBySlug } from "@/features/events/queries";
import { formatDateTime } from "@/lib/format";
import { getEventImageUrl } from "@/lib/storage/helpers";

export const dynamic = "force-dynamic";

export default async function EventDetailPage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const event = await getPublishedEventBySlug(slug);
  if (!event) notFound();

  const imageUrl = getEventImageUrl(event.image_path);

  return (
    <Box sx={{ maxWidth: 760, mx: "auto" }}>
      <Button href="/eventos" sx={{ mb: 2 }}>
        ← Voltar para eventos
      </Button>
      <Typography variant="h4" component="h1" sx={{ mb: 1 }}>
        {event.title}
      </Typography>
      <Typography variant="subtitle1" color="text.secondary" sx={{ mb: 2 }}>
        {formatDateTime(event.starts_at)}
        {event.ends_at && ` até ${formatDateTime(event.ends_at)}`}
      </Typography>
      {event.location && (
        <Typography variant="body1" sx={{ mb: 2 }}>
          📍 {event.location}
        </Typography>
      )}
      {imageUrl && (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          src={imageUrl}
          alt={event.title}
          style={{ width: "100%", borderRadius: 12, marginBottom: 16 }}
        />
      )}
      {event.description && (
        <Typography
          variant="body1"
          sx={{ whiteSpace: "pre-wrap", lineHeight: 1.7 }}
        >
          {event.description}
        </Typography>
      )}
    </Box>
  );
}