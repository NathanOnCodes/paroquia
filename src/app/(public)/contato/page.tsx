import Box from "@mui/material/Box";
import PageHeader from "@/design-system/layout/PageHeader";
import ContactForm from "@/features/contact/components/ContactForm";

export const metadata = {
  title: "Contato",
};

export default function ContatoPage() {
  return (
    <Box sx={{ maxWidth: 640, mx: "auto" }}>
      <PageHeader
        title="Fale conosco"
        description="Envie sua mensagem: dúvidas, sugestões ou pedidos de oração."
      />
      <ContactForm />
    </Box>
  );
}