import { notFound } from "next/navigation";
import Typography from "@mui/material/Typography";
import TextField from "@mui/material/TextField";
import Button from "@mui/material/Button";
import ServerForm from "@/components/admin/ServerForm";
import { getCommunityById } from "@/features/communities/queries.admin";
import { updateCommunityAction } from "@/features/communities/actions";

export const dynamic = "force-dynamic";

export default async function CommunityEditPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const community = await getCommunityById(id);
  if (!community) notFound();

  return (
    <>
      <Button href="/admin/comunidades" sx={{ mb: 2 }}>
        ← Voltar
      </Button>
      <Typography variant="h5" sx={{ mb: 2 }}>
        Editar comunidade
      </Typography>
      <ServerForm action={updateCommunityAction} cancelHref="/admin/comunidades">
        <input type="hidden" name="id" value={community.id} />
        <TextField label="Nome" name="name" defaultValue={community.name} required fullWidth />
        <TextField label="Descrição" name="description" defaultValue={community.description ?? ""} multiline rows={2} fullWidth />
        <TextField label="Cidade" name="city" defaultValue={community.city ?? ""} fullWidth />
      </ServerForm>
    </>
  );
}