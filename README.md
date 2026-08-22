# Sistema para Igrejas: dízimos, doações e eventos

Este projeto nasceu com um propósito simples: contribuir com a Santa Igreja
Católica oferecendo um sistema gratuito para gerenciamento de dizimistas e
doações, publicação de eventos e transparência financeira, para que as
paróquias possam se dedicar ao que realmente importa: a missão.

A ideia foi desenvolvida como um piloto para a **Paróquia São Benedito e
Menino Jesus**, em Francisco Morato - SP, mas o sistema foi feito para ser
**reutilizado por qualquer paróquia, capela, comunidade ou movimento da
Igreja**: pode ser copiado, adaptado e modificado livremente para atender a
realidade de cada comunidade.

## O que o sistema faz

- **Dízimo online**: valor livre, avulso ou recorrente (PIX, cartão de
  débito e crédito via Stripe), com recibo por e-mail.
- **Comunidades**: o fiel indica a comunidade da qual participa ou digita o
  nome dela quando não está na lista (inclusive de outra cidade).
- **Gerenciamento de dizimistas**: cadastro, histórico de contribuições,
  recorrências e cancelamento feito pelo próprio fiel, sem senha.
- **Eventos**: publicação pública com data, horário, local e imagem.
- **Transparência financeira**: períodos com receitas, despesas, saldo e
  documentos DRE em PDF ou imagem, publicados pelo pároco.
- **Painel administrativo**: dashboard com ranking das comunidades que mais
  contribuem, mensagens de contato e gestão de usuários (pároco e
  assistentes).

## Como usar em sua paróquia

Todo o código é aberto. Você precisa apenas de contas gratuitas no Supabase
(banco de dados), Stripe (pagamentos) e Resend (e-mails) e seguir o passo a
passo abaixo. Nenhum dado dos fiéis sai do seu próprio banco.

## Stack

- **Next.js 16** (App Router) + TypeScript
- **Supabase**: PostgreSQL, Auth, Storage, RLS
- **Stripe**: dízimo online (PIX, cartão de crédito/débito, recorrência)
- **Resend**: recibos e e-mails transacionais
- **Material UI** + Tailwind CSS
- **Vercel**: deploy

## Começando

### 1. Prerequisitos

- Node.js 20+
- Conta no [Supabase](https://supabase.com)
- Conta no [Stripe](https://stripe.com)
- Conta no [Resend](https://resend.com)
- Supabase CLI (`npm i -g supabase`) (opcional para rodar banco local)

### 2. Banco de dados

1. Crie um projeto no Supabase.
2. Aplique as migrações `supabase/migrations/00001_init.sql` e
   `supabase/migrations/00002_comunidade_livre.sql` no **SQL Editor**
   do Supabase, nesta ordem (ou via CLI: `supabase db push`).
3. As migrações criam tabelas, políticas RLS, buckets de storage
   (`events` público e `transparency` privado) e as comunidades iniciais.
4. No SQL Editor, ajuste os nomes das comunidades para as da sua paróquia
   (`update communities set name = '...'` ou cadastre pelo painel).

### 3. Variáveis de ambiente

Copie `.env.example` para `.env.local` e preencha:

| Variável | Onde obter |
| --- | --- |
| `NEXT_PUBLIC_SUPABASE_URL` | Configurações do Supabase (Project Settings → API) |
| `NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY` | Chave `anon`/`publishable` |
| `SUPABASE_SERVICE_ROLE_KEY` | Chave `service_role` (só no servidor, nunca expor) |
| `NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY` | Stripe Dashboard → Developers |
| `STRIPE_SECRET_KEY` | Stripe Dashboard |
| `STRIPE_WEBHOOK_SECRET` | Stripe → Webhooks (ver abaixo) |
| `RESEND_API_KEY` | Resend Dashboard |
| `EMAIL_FROM` | Endereço remetente validado no Resend |
| `NEXT_PUBLIC_SITE_URL` | URL do site |

### 4. Stripe

1. Habilite PIX e cartões na sua conta Stripe (Brasil) conforme a
   disponibilidade. Para recorrência via PIX, habilite o **Pix Automático**.
2. Crie um webhook em **Developers → Webhooks** apontando para
   `https://SEU-SITE/api/stripe/webhook` com os eventos:
   - `payment_intent.succeeded`, `payment_intent.payment_failed`,
     `payment_intent.canceled`, `charge.refunded`
   - `invoice.payment_succeeded`, `invoice.payment_failed`
   - `customer.subscription.created`, `customer.subscription.updated`,
     `customer.subscription.deleted`
3. Copie o valor `whsec_...` para `STRIPE_WEBHOOK_SECRET`.

### 5. Usuário administrador

Crie o primeiro usuário administrativo manualmente no Supabase
(**Authentication → Users → Add user**), com e-mail/senha. Depois insira na
tabela `profiles` (via SQL Editor):

```sql
insert into profiles (id, full_name, email, role, is_active)
select id, 'Seu Nome', email, 'admin', true
from auth.users
where email = 'SEU-EMAIL';
```

Somente um usuário com papel `admin` (na interface exibido como "Pároco")
pode criar assistentes e publicar a transparência financeira. Os nomes
exibidos podem ser adaptados na interface sem alterar o código.

### 6. Rodando

```bash
npm install
npm run dev
```

## Comandos

```bash
npm run dev           # desenvolvimento
npm run build         # build de produção
npm run start         # servidor de produção
npm run lint          # eslint
npm run typecheck     # verificação de tipos
npm run test:security # testes de segurança (vitest)
```

## Deploy na Vercel

1. Suba o projeto para um repositório GitHub.
2. Importe no Vercel (framework: Next.js).
3. Configure as variáveis de ambiente de **produção** no painel da Vercel.
4. Adicione o webhook da Stripe com a URL de produção.

## Segurança e privacidade

- Todas as rotas `/admin` exigem autenticação e verificam o papel no servidor.
- O Supabase aplica **RLS**: dados de dizimistas e finanças não são legíveis
  publicamente.
- A chave `service_role` fica somente no servidor.
- Documentos de transparência são servidos via URL assinada apenas quando o
  período está publicado.
- Tokens de gerenciamento de recorrência são armazenados apenas com hash.
- Nenhum dado pessoal é exposto no site público; a prestação de contas mostra
  apenas valores agregados.

## Licença e reutilização

Este projeto é distribuído sob a **Licença de Caridade para a Igreja**
(veja o arquivo `LICENSE`), criada para manter o sistema como um dom:

- **Permitido**: usar, copiar, estudar, modificar e adaptar livremente em
  benefício da Igreja.
- **Obrigatório**: versões modificadas seguem a mesma licença gratuita e
  podem ser doadas a outras comunidades que as solicitarem.
- **Proibido**: vender o sistema, cobrar por ele como produto ou serviço,
  ou usá-lo com fins comerciais lucrativos. Apenas custos diretos de
  infraestrutura (hospedagem, domínio, pagamentos) podem ser reembolsados,
  sem lucro.

## Autor

Desenvolvido com fé por **Nathan**, como obra de caridade a serviço da
Santa Igreja.

- LinkedIn: [linkedin.com/in/nathanoncodes](https://www.linkedin.com/in/nathanoncodes/)

> "Cada um contribua segundo propôs no seu coração, não com tristeza ou por
> necessidade; porque Deus ama ao que dá com alegria." (2 Coríntios 9, 7)
