import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Link from "@mui/material/Link";
import BrandMark from "@/design-system/branding/BrandMark";
import Surface from "@/design-system/layout/Surface";
import LoginForm from "@/features/auth/components/LoginForm";

export const metadata = {
  title: "Acesso restrito",
};

export default function LoginPage() {
  return (
    <Box
      sx={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        bgcolor: "background.default",
        p: 2,
      }}
    >
      <Surface sx={{ maxWidth: 400, width: "100%", bgcolor: "white" }}>
        <Box sx={{ display: "flex", flexDirection: "column", alignItems: "center", mb: 3 }}>
          <BrandMark size={56} sx={{ mb: 2 }} />
          <Typography variant="h5" sx={{ fontWeight: 700 }}>
            Acesso restrito
          </Typography>
          <Typography
            variant="body2"
            color="text.secondary"
            sx={{ mt: 1, textAlign: "center", fontStyle: "italic", maxWidth: 320 }}
          >
            &ldquo;Depois de Deus, o sacerdote é tudo!&rdquo;
            <br />
            <Box component="span" sx={{ fontSize: "0.75rem", opacity: 0.7 }}>
              — São João Maria Vianney
            </Box>
          </Typography>
        </Box>
        <LoginForm />
        <Typography variant="body2" sx={{ mt: 3, textAlign: "center" }}>
          <Link href="/" underline="hover">
            ← Voltar ao site
          </Link>
        </Typography>
      </Surface>
    </Box>
  );
}