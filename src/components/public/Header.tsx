import Link from "next/link";
import AppBar from "@mui/material/AppBar";
import Toolbar from "@mui/material/Toolbar";
import Button from "@mui/material/Button";
import Box from "@mui/material/Box";
import Container from "@mui/material/Container";

export default function PublicHeader() {
  return (
    <AppBar position="static" color="primary" elevation={0}>
      <Toolbar sx={{ minHeight: { xs: 56, sm: 64 } }}>
        <Container
          maxWidth="lg"
          sx={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: 2, px: { xs: 1, sm: 2 } }}
        >
          <Link href="/" style={{ textDecoration: "none", color: "inherit" }}>
            <Box component="span" sx={{ fontSize: { xs: 15, sm: 20 }, fontWeight: 700, lineHeight: 1.15, display: "block", maxWidth: { xs: 190, sm: 360 } }}>
              Paróquia São Benedito e Menino Jesus
            </Box>
          </Link>
          <Box sx={{ display: "flex", gap: { xs: 0.5, sm: 1 }, flexWrap: "wrap" }}>
            <Button href="/eventos" color="inherit">
              Eventos
            </Button>
            <Button href="/dizimo" color="inherit">
              Dízimo
            </Button>
            <Button href="/transparencia" color="inherit">
              Transparência
            </Button>
            <Button href="/contato" color="inherit">
              Contato
            </Button>
          </Box>
        </Container>
      </Toolbar>
    </AppBar>
  );
}
