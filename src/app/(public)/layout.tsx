import Box from "@mui/material/Box";
import Container from "@mui/material/Container";
import PublicHeader from "@/components/public/Header";
import PublicFooter from "@/components/public/Footer";
import { getSiteSettings } from "@/features/settings/queries";

export default async function PublicLayout({ children }: { children: React.ReactNode }) {
  const settings = await getSiteSettings();

  return (
    <Box sx={{ display: "flex", flexDirection: "column", minHeight: "100vh", bgcolor: "background.default" }}>
      <PublicHeader settings={settings} />
      <Box component="main" sx={{ flex: 1 }}>
        <Container maxWidth="lg" sx={{ py: { xs: 12, md: 6 } }}>
          {children}
        </Container>
      </Box>
      <PublicFooter settings={settings} />
    </Box>
  );
}
