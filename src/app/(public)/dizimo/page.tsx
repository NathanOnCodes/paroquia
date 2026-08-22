import Typography from "@mui/material/Typography";
import Box from "@mui/material/Box";
import TitheForm from "@/components/payments/TitheForm";
import { listActiveCommunities } from "@/features/communities/queries";

export const dynamic = "force-dynamic";

export const metadata = {
  title: "Dízimo online",
  description:
    "Contribua com o dízimo de maneira avulsa ou recorrente, via PIX ou cartão.",
};

export default async function DizimoPage() {
  const communities = await listActiveCommunities();

  return (
    <Box sx={{ maxWidth: 640, mx: "auto" }}>
      <Typography variant="h4" sx={{ mb: 1 }}>
        Dízimo online
      </Typography>
      <Typography color="text.secondary" sx={{ mb: 3 }}>
        Sua contribuição sustenta as atividades da paróquia e das comunidades.
        Você pode doar uma vez ou tornar seu dízimo recorrente todo mês.
      </Typography>
      <TitheForm communities={communities} />
    </Box>
  );
}
