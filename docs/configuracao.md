# Configuração

O site funciona com uma identidade genérica até que um administrador configure
os dados da paróquia em `/admin/configuracoes`. A organização do código está
em `docs/arquitetura.md`.

## Teste local rápido

Em desenvolvimento, o projeto entra automaticamente em modo de demonstração
quando as variáveis de infraestrutura não estão preenchidas. Nesse modo:

- A home abre com os valores genéricos.
- Eventos e comunidades aparecem vazios, sem exigir Supabase local.
- Pagamentos reais ficam desativados.
- O formulário recorrente pode ser visualizado, mas não processa cobranças.

Execute apenas `npm run dev` e abra `http://localhost:3000`. Para usar banco,
pagamentos e e-mails reais, preencha as variáveis abaixo e reinicie o servidor.

## Primeiro acesso

1. Configure as variáveis obrigatórias no ambiente de execução.
2. Execute as migrations do Supabase, na ordem:
   `00001_init.sql`, `00002_comunidade_livre.sql` e `00003_site_settings.sql`.
3. Entre com um usuário `admin`.
4. Abra `/admin/configuracoes` e salve o nome, a cidade, a cor e os dados Pix.

Nome e cor não são obrigatórios no ambiente. Sem configuração, o site usa:

- Nome: `Paróquia`
- Cor secundária: `#1D4E89` (`rgb(29, 78, 137)`), um azul mariano escuro
- Fundo: `#F8F9FA`

## Rotas públicas

- `/` home com identidade configurada
- `/eventos` e `/eventos/[slug]`
- `/doacoes` (Pix + doação recorrente) — `/dizimo` redireciona para cá
- `/doacoes/gerenciar` cancelamento pelo fiel
- `/transparencia` e `/transparencia/[id]`
- `/contato` e `/privacidade`

## Variáveis de ambiente

O prefixo `NEXT_PUBLIC_` é exigido pelo Next.js para valores que precisam ser
lidos no navegador. Chaves sem esse prefixo permanecem somente no servidor.

| Variável | Obrigatória | Uso |
| --- | --- | --- |
| `NEXT_PUBLIC_URL_SUPABASE` | Sim | URL pública do projeto Supabase |
| `NEXT_PUBLIC_CHAVE_PUBLICA_SUPABASE` | Sim | Chave pública do Supabase |
| `CHAVE_SERVICO_SUPABASE` | Sim | Chave privada usada somente no servidor |
| `NEXT_PUBLIC_CHAVE_PUBLICA_STRIPE` | Sim | Chave pública do Stripe |
| `CHAVE_SECRETA_STRIPE` | Sim | Chave privada do Stripe |
| `SEGREDO_WEBHOOK_STRIPE` | Sim | Validação dos webhooks Stripe |
| `CHAVE_API_RESEND` | Sim | Envio de e-mails |
| `EMAIL_REMETENTE` | Não | Remetente dos e-mails |
| `EMAIL_NOTIFICACAO_ADMIN` | Não | Endereço que recebe alertas de contato |
| `NEXT_PUBLIC_URL_SITE` | Não | URL pública, padrão `http://localhost:3000` |
| `TTL_HORAS_TOKEN_RECORRENTE` | Não | Validade do link de recorrência |
| `TAMANHO_MAXIMO_ARQUIVO_MB` | Não | Limite de arquivos enviados |

Não coloque chaves reais, `CHAVE_SERVICO_SUPABASE`, `CHAVE_SECRETA_STRIPE` ou
`SEGREDO_WEBHOOK_STRIPE` no repositório. O arquivo `.env.local` não deve ser
versionado.

## Pix

O QR Code Pix é estático e de valor aberto. Informe no painel uma chave Pix
válida, o nome do recebedor e a cidade. O sistema gera o QR Code e o código
Pix copia e cola automaticamente usando o padrão BR Code.

A chave pode ser um e-mail, CNPJ, CPF, telefone ou chave aleatória, mas precisa
estar cadastrada na conta bancária recebedora. O QR estático não confirma o
pagamento nem gera conciliação automática.

## Migrando variáveis antigas

| Nome antigo | Nome atual |
| --- | --- |
| `NEXT_PUBLIC_SUPABASE_URL` | `NEXT_PUBLIC_URL_SUPABASE` |
| `NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY` | `NEXT_PUBLIC_CHAVE_PUBLICA_SUPABASE` |
| `SUPABASE_SERVICE_ROLE_KEY` | `CHAVE_SERVICO_SUPABASE` |
| `NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY` | `NEXT_PUBLIC_CHAVE_PUBLICA_STRIPE` |
| `STRIPE_SECRET_KEY` | `CHAVE_SECRETA_STRIPE` |
| `STRIPE_WEBHOOK_SECRET` | `SEGREDO_WEBHOOK_STRIPE` |
| `RESEND_API_KEY` | `CHAVE_API_RESEND` |
| `EMAIL_FROM` | `EMAIL_REMETENTE` |
| `ADMIN_NOTIFICATION_EMAIL` | `EMAIL_NOTIFICACAO_ADMIN` |
| `NEXT_PUBLIC_SITE_URL` | `NEXT_PUBLIC_URL_SITE` |
| `RECURRING_TOKEN_TTL_HOURS` | `TTL_HORAS_TOKEN_RECORRENTE` |
| `MAX_FILE_SIZE_MB` | `TAMANHO_MAXIMO_ARQUIVO_MB` |
