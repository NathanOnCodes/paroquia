import { redirect } from "next/navigation";
import Box from "@mui/material/Box";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import Divider from "@mui/material/Divider";
import Image from "next/image";
import ServerForm from "@/design-system/forms/ServerForm";
import Surface from "@/design-system/layout/Surface";
import PageHeader from "@/design-system/layout/PageHeader";
import { getSiteSettings } from "@/features/settings/queries";
import { updateSiteSettingsAction } from "@/features/settings/actions";
import { getAdminUser } from "@/lib/auth/authorization";
import { generateOpenPixQr } from "@/features/payments/pix";

export const dynamic = "force-dynamic";

export default async function SettingsPage() {
  const admin = await getAdminUser("admin");
  if (!admin) redirect("/admin");

  const settings = await getSiteSettings();
  const pixPreview = await generateOpenPixQr(settings);

  return (
    <Box sx={{ maxWidth: 900 }}>
      <PageHeader
        title="Configurações do site"
        description="Personalize a identidade da paróquia sem precisar fazer um novo deploy. Todos os campos são opcionais para o site funcionar."
      />
      <Surface>
        <ServerForm action={updateSiteSettingsAction} successMessage="Configurações atualizadas!">
          <Typography variant="h6">Identidade</Typography>
          <Box sx={{ display: "grid", gap: 2, gridTemplateColumns: { xs: "1fr", sm: "1fr 1fr" } }}>
            <TextField name="site_name" label="Nome da paróquia" defaultValue={settings.site_name} required />
            <TextField name="site_title" label="Título do site" defaultValue={settings.site_title} required />
            <TextField name="city" label="Cidade" defaultValue={settings.city} />
            <TextField name="secondary_color" label="Cor secundária" defaultValue={settings.secondary_color} helperText="Formato hexadecimal, por exemplo #1D4E89" required />
          </Box>
          <TextField name="description" label="Descrição curta" defaultValue={settings.description} required />
          <Divider />
          <Typography variant="h6">Pix opcional</Typography>
          <Typography variant="body2" color="text.secondary">
            Informe uma chave Pix cadastrada na conta da paróquia para gerar um QR Code estático de valor aberto.
          </Typography>
          <TextField name="pix_key" label="Chave Pix (e-mail, CNPJ, telefone ou chave aleatória)" defaultValue={settings.pix_key ?? ""} />
          <Box sx={{ display: "grid", gap: 2, gridTemplateColumns: { xs: "1fr", sm: "1fr 1fr" } }}>
            <TextField name="pix_receiver_name" label="Nome do recebedor Pix" defaultValue={settings.pix_receiver_name ?? ""} helperText="Até 25 caracteres" />
            <TextField name="pix_receiver_city" label="Cidade do recebedor Pix" defaultValue={settings.pix_receiver_city ?? ""} helperText="Até 15 caracteres" />
          </Box>
          <Divider />
          <Typography variant="h6">Home</Typography>
          <TextField name="home_title" label="Título personalizado (opcional)" defaultValue={settings.home_title ?? ""} />
          <TextField name="home_description" label="Descrição personalizada (opcional)" defaultValue={settings.home_description ?? ""} multiline minRows={3} />
          <Box sx={{ display: "flex", alignItems: "center", gap: 2, flexWrap: "wrap", p: 2, borderRadius: 3, bgcolor: settings.secondary_color, color: "secondary.contrastText" }}>
            <Box>
              <Typography variant="subtitle2">Prévia da identidade</Typography>
              <Typography variant="h6">{settings.site_name}</Typography>
              <Typography variant="body2">Fundo marfim com cor secundária configurável.</Typography>
            </Box>
            {pixPreview && <Image src={pixPreview.qrCodeImage} alt="Prévia do QR Code Pix" width={96} height={96} unoptimized />}
          </Box>
        </ServerForm>
      </Surface>
    </Box>
  );
}
