import Typography from "@mui/material/Typography";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Alert from "@mui/material/Alert";

export const metadata = {
  title: "Dízimo enviado",
};

export default function DizimoSuccessPage() {
  return (
    <Box sx={{ maxWidth: 560, mx: "auto", textAlign: "center" }}>
      <Alert severity="success" sx={{ mb: 3 }}>
        Pagamento recebido!
      </Alert>
      <Typography variant="h5" sx={{ mb: 2 }}>
        Obrigado pela sua contribuição.
      </Typography>
      <Typography color="text.secondary" sx={{ mb: 3 }}>
        O recibo será enviado por e-mail assim que o pagamento for confirmado.
        Que Deus abençoe você e sua família.
      </Typography>
      <Box sx={{ display: "flex", gap: 2, justifyContent: "center", flexWrap: "wrap" }}>
        <Button href="/" variant="contained">
          Voltar ao início
        </Button>
        <Button href="/dizimo/gerenciar" variant="outlined">
          Gerenciar meu dízimo recorrente
        </Button>
      </Box>
    </Box>
  );
}