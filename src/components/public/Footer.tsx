import Link from "next/link";
import Box from "@mui/material/Box";
import Container from "@mui/material/Container";
import Typography from "@mui/material/Typography";

export default function PublicFooter() {
  return (
    <Box component="footer" sx={{ bgcolor: "#0f2440", color: "#fff", mt: 8, py: 4 }}>
      <Container maxWidth="lg">
        <Box sx={{ display: "flex", justifyContent: "space-between", flexWrap: "wrap", gap: 2 }}>
          <Box>
            <Typography variant="h6" sx={{ fontWeight: 700, maxWidth: 360 }}>
              Paróquia São Benedito e Menino Jesus
            </Typography>
            <Typography variant="body2" sx={{ opacity: 0.8 }}>
              Francisco Morato - SP
            </Typography>
            <Typography variant="body2" sx={{ opacity: 0.8, mt: 0.5 }}>
              Comunidade, fé e transparência.
            </Typography>
          </Box>
          <Box sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
            <Link href="/eventos" style={{ color: "inherit", textDecoration: "none" }}>
              Eventos
            </Link>
            <Link href="/dizimo" style={{ color: "inherit", textDecoration: "none" }}>
              Dízimo online
            </Link>
            <Link href="/transparencia" style={{ color: "inherit", textDecoration: "none" }}>
              Transparência financeira
            </Link>
            <Link href="/contato" style={{ color: "inherit", textDecoration: "none" }}>
              Contato
            </Link>
            <Link href="/privacidade" style={{ color: "inherit", textDecoration: "none" }}>
              Política de privacidade
            </Link>
          </Box>
        </Box>
      </Container>
    </Box>
  );
}
