import Box from "@mui/material/Box";
import Paper from "@mui/material/Paper";
import Typography from "@mui/material/Typography";
import Link from "@mui/material/Link";
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
      <Paper sx={{ p: 4, maxWidth: 400, width: "100%" }} elevation={3}>
        <Typography variant="h5" sx={{ mb: 0.5, textAlign: "center" }}>
          Acesso restrito
        </Typography>
        <Typography color="text.secondary" sx={{ mb: 3, textAlign: "center" }}>
          Área exclusiva do pároco e seus assistentes.
        </Typography>
        <LoginForm />
        <Typography variant="body2" sx={{ mt: 3, textAlign: "center" }}>
          <Link href="/" underline="hover">
            ← Voltar ao site
          </Link>
        </Typography>
      </Paper>
    </Box>
  );
}