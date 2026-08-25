import Link from "next/link";
import Box from "@mui/material/Box";
import Container from "@mui/material/Container";
import Typography from "@mui/material/Typography";
import type { SiteSettings } from "@/features/settings/types";

export default function PublicFooter({ settings }: { settings: SiteSettings }) {
  return (
    <Box component="footer" sx={{ mt: 8, py: 4, borderTop: 1, borderColor: "divider" }}>
      <Container maxWidth="lg">
        <Box sx={{ display: "flex", justifyContent: "space-between", flexWrap: "wrap", gap: 2 }}>
          <Box>
            <Typography variant="h6" sx={{ fontWeight: 700, maxWidth: 360 }}>
              {settings.site_name}
            </Typography>
            <Typography variant="body2" sx={{ opacity: 0.8 }}>
              {settings.city}
            </Typography>
            <Typography variant="body2" sx={{ opacity: 0.8, mt: 0.5 }}>
              Comunidade, fé e transparência.
            </Typography>
          </Box>
          <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
            <Link href="/eventos" style={{ textDecoration: "none" }}>
              Eventos
            </Link>
            <Link href="/doacoes" style={{ textDecoration: "none" }}>
              Doações online
            </Link>
            <Link href="/transparencia" style={{ textDecoration: "none" }}>
              Transparência financeira
            </Link>
            <Link href="/contato" style={{ textDecoration: "none" }}>
              Contato
            </Link>
            <Link href="/privacidade" style={{ textDecoration: "none" }}>
              Política de privacidade
            </Link>
          </Box>
        </Box>
      </Container>
    </Box>
  );
}
