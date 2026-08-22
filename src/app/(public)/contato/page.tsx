import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import ContactForm from "@/components/contact/ContactForm";

export const metadata = {
  title: "Contato",
};

export default function ContatoPage() {
  return (
    <Box sx={{ maxWidth: 640, mx: "auto" }}>
      <Typography variant="h4" sx={{ mb: 1 }}>
        Fale conosco
      </Typography>
      <Typography color="text.secondary" sx={{ mb: 3 }}>
        Envie sua mensagem: dúvidas, sugestões ou pedidos de oração.
      </Typography>
      <ContactForm />
    </Box>
  );
}