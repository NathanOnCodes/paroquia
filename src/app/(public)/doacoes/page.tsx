import Box from "@mui/material/Box";
import TitheForm from "@/features/payments/components/TitheForm";
import DonationTabs from "@/features/payments/components/DonationTabs";
import PixDonationPanel from "@/features/payments/components/PixDonationPanel";
import { listActiveCommunities } from "@/features/communities/queries";
import { getSiteSettings } from "@/features/settings/queries";
import { generateOpenPixQr } from "@/features/payments/pix";
import PageHeader from "@/design-system/layout/PageHeader";

export const dynamic = "force-dynamic";

export const metadata = {
  title: "Doações",
  description:
    "Contribua com a paróquia via Pix ou doação recorrente.",
};

export default async function DonationsPage() {
  const [communities, settings] = await Promise.all([listActiveCommunities(), getSiteSettings()]);
  const pixPayment = await generateOpenPixQr(settings);

  return (
    <Box sx={{ maxWidth: 760, mx: "auto" }}>
      <PageHeader
        title="Doações"
        description="Sua contribuição ajuda a manter a comunidade, as atividades e a missão da paróquia."
      />
      <DonationTabs
        pixPanel={<PixDonationPanel payment={pixPayment} />}
        recurringPanel={<TitheForm communities={communities} recurringOnly />}
      />
    </Box>
  );
}
