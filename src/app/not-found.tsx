import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";

export const dynamic = "force-static";

export default function NotFound() {
  return (
    <Box
      sx={{
        minHeight: "60vh",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        textAlign: "center",
        gap: 2,
        py: 8,
      }}
    >
      <Typography variant="h1" sx={{ color: "primary.main", fontSize: { xs: 72, md: 96 } }}>
        404
      </Typography>
      <Typography variant="h4" component="h1">
        Página não encontrada
      </Typography>
      <Typography color="text.secondary" sx={{ maxWidth: 420 }}>
        A página que você procura não existe ou foi movida. Verifique o endereço
        ou volte para o início.
      </Typography>
      <Button href="/" variant="contained" size="large" sx={{ mt: 2 }}>
        Voltar ao início
      </Button>
    </Box>
  );
}