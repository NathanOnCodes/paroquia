import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";

export const metadata = {
  title: "Política de privacidade",
};

export default function PrivacidadePage() {
  return (
    <Box sx={{ maxWidth: 720, mx: "auto" }}>
      <Typography variant="h4" sx={{ mb: 3 }}>
        Política de privacidade
      </Typography>
      <Typography variant="body1" sx={{ whiteSpace: "pre-wrap", lineHeight: 1.8 }} component="div">
        {`A paróquia respeita a sua privacidade. Este documento explica como tratamos os dados pessoais fornecidos neste site.

Dados coletados:
- Formulário de contato: nome, e-mail, telefone (opcional) e mensagem.
- Dízimo online: nome, e-mail, telefone, cidade, bairro (opcional) e comunidade.

Finalidade:
- Gerenciar as contribuições de dízimo e emitir recibos.
- Responder mensagens enviadas pelo formulário de contato.
- Elaborar relatórios agregados e a prestação de contas (sem dados pessoais).

Pagamentos:
- Os dados de pagamento (cartão, PIX) são processados exclusivamente pela Stripe.
- Não armazenamos números de cartão, códigos de segurança ou dados equivalentes.

Compartilhamento:
- Não vendemos nem compartilhamos seus dados com terceiros.
- Dados podem ser acessados apenas pela equipe autorizada da paróquia.

Direitos:
- Você pode solicitar acesso, correção ou exclusão dos seus dados pessoais
  entrando em contato pelo formulário do site ou pelo e-mail administrativo.

Consentimento:
- Ao utilizar o dízimo online ou o formulário de contato, você consente com o
  tratamento descrito nesta política.

Armazenamento:
 - Os dados são armazenados com segurança, têm acesso restrito e são protegidos por
  políticas de segurança de nível de linha no banco de dados.`}
      </Typography>
    </Box>
  );
}
