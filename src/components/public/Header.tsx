import Link from "next/link";
import AppBar from "@mui/material/AppBar";
import Toolbar from "@mui/material/Toolbar";
import Button from "@mui/material/Button";
import Box from "@mui/material/Box";
import Container from "@mui/material/Container";
import type { SiteSettings } from "@/features/settings/types";
import BrandMark from "@/design-system/branding/BrandMark";

export default function PublicHeader({ settings }: { settings: SiteSettings }) {
  return (
    <AppBar position="static" color="transparent" elevation={0} sx={{ bgcolor: "transparent", color: "text.primary", boxShadow: "none", border: 0 }}>
      <Toolbar sx={{ minHeight: { xs: 56, sm: 64 } }}>
        <Container
          maxWidth="lg"
          sx={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: 2, px: { xs: 1, sm: 2 } }}
        >
          <Link href="/" style={{ textDecoration: "none" }}>
            <Box sx={{ display: "flex", alignItems: "center", gap: 1, color: "text.primary", "&:hover": { color: "secondary.main" } }}>
              <BrandMark />
              <Box component="span" sx={{ fontSize: { xs: 15, sm: 20 }, fontWeight: 700, lineHeight: 1.15, display: "block", maxWidth: { xs: 190, sm: 360 } }}>
                {settings.site_name}
              </Box>
            </Box>
          </Link>
          <Box sx={{ display: "flex", gap: { xs: 0.5, sm: 1 }, flexWrap: "wrap" }}>
            <Button href="/eventos" color="inherit" sx={{ "&:hover": { color: "secondary.main", bgcolor: "transparent" } }}>
              Eventos
            </Button>
            <Button href="/doacoes" color="inherit" sx={{ "&:hover": { color: "secondary.main", bgcolor: "transparent" } }}>
              Doações
            </Button>
            <Button href="/transparencia" color="inherit" sx={{ "&:hover": { color: "secondary.main", bgcolor: "transparent" } }}>
              Transparência
            </Button>
            <Button href="/contato" color="inherit" sx={{ "&:hover": { color: "secondary.main", bgcolor: "transparent" } }}>
              Contato
            </Button>
          </Box>
        </Container>
      </Toolbar>
    </AppBar>
  );
}
