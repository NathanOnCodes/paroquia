import Typography from "@mui/material/Typography";
import TextField from "@mui/material/TextField";
import MenuItem from "@mui/material/MenuItem";
import ServerForm from "@/design-system/forms/ServerForm";
import { createDonorAction } from "@/features/donors/actions";
import { listAllCommunities } from "@/features/communities/queries.admin";

export const dynamic = "force-dynamic";

export default async function NewDonorPage() {
  const communities = await listAllCommunities();

  return (
    <>
      <Typography variant="h5" sx={{ mb: 2 }}>
        Novo dizimista
      </Typography>
      <ServerForm action={createDonorAction} cancelHref="/admin/dizimistas" submitLabel="Cadastrar">
        <TextField label="Nome completo" name="full_name" required fullWidth />
        <TextField label="E-mail" name="email" type="email" required fullWidth />
        <TextField label="Telefone" name="phone" fullWidth placeholder="(00) 00000-0000" />
        <TextField select label="Comunidade" name="community_id" fullWidth defaultValue="">
          <MenuItem value="">Sem comunidade</MenuItem>
          {communities.map((c) => (
            <MenuItem key={c.id} value={c.id}>
              {c.name}
            </MenuItem>
          ))}
        </TextField>
        <TextField
          label="Comunidade fora da lista (opcional)"
          name="community_name"
          helperText="Use quando o dizimista pertence a outra paróquia ou cidade."
          fullWidth
        />
        <TextField label="Cidade" name="city" fullWidth />
        <TextField label="Bairro" name="neighborhood" fullWidth />
      </ServerForm>
    </>
  );
}