# Arquitetura do projeto

Este documento define onde cada coisa deve ficar no código. Siga estas
regras ao criar ou mover arquivos, para que o projeto continue organizado
conforme cresce.

## Visão geral

O projeto usa **Vertical Slice Architecture** com um **design system
compartilhado**: cada funcionalidade (slice) concentra suas próprias regras,
e o visual padronizado fica em um lugar único, usado por todas as páginas.

```text
src/
  app/            rotas, layouts e handlers (só composição, sem regra de negócio)
  components/     layouts compartilhados (Header, Footer, AdminShell)
  design-system/  componentes visuais universais + tema MUI
  features/       vertical slices (regras de negócio por funcionalidade)
  lib/            infraestrutura transversal (Supabase, Stripe, e-mail, auth)
  tests/          testes de segurança (vitest)
supabase/
  migrations/     SQL versionado do banco
docs/             documentação do projeto
```

## Onde colocar cada coisa

### `src/app` — rotas

Somente páginas, layouts, `route.ts` e arquivos especiais do Next.js
(`error.tsx`, `not-found.tsx`, `proxy.ts` na raiz de `src`).

- Páginas **públicas** ficam em `src/app/(public)/...`.
- Páginas **administrativas** ficam em `src/app/admin/...`.
- **APIs** ficam em `src/app/api/...`.

Regra: uma página em `src/app` deve ser fina. Ela busca dados chamando
queries da feature e renderiza componentes. Se houver regra de negócio,
ela pertence à feature.

### `src/features/<slice>` — funcionalidades

Cada pasta é uma funcionalidade completa. Slices atuais: `auth`,
`communities`, `contact`, `contributions`, `dashboard`, `donors`, `events`,
`payments`, `recurring`, `settings`, `transparency`.

Estrutura padrão de uma slice:

```text
features/<slice>/
  actions.ts      server actions (mutações, com "use server")
  queries.ts      consultas de leitura (públicas ou admin)
  queries.admin.ts  consultas administrativas (opcional)
  schemas.ts      validação Zod
  services.ts     integrações externas (Stripe, e-mail, tokens)
  types.ts        tipos do domínio (opcional)
  components/     componentes React específicos desta funcionalidade
  webhook.ts      processamento de webhooks (apenas payments)
```

Convenções:

- Uma slice **não importa queries de outra slice** por caminho errado.
  Se duas slices precisam dos mesmos dados, uma chama um serviço da outra
  de forma explícita (ex.: `recurring` chama serviços de `payments`) ou os
  dados vêm por props.
- `payments` é dona do Stripe (PaymentIntent, SetupIntent, webhook,
  persistência de `contribution_payments`) e do **Pix estático** (`pix.ts`).
- `recurring` é dona da experiência de recorrência do fiel (link mágico,
  cancelamento) e chama `payments` quando precisa tocar a Stripe.
- `settings` é dona da identidade da paróquia (nome, cor, Pix) na tabela
  `site_settings`; o tema e os layouts leem dela.
- Queries administrativas de uma entidade ficam na slice da entidade
  (ex.: períodos financeiros em `transparency`, comunidades em
  `communities/queries.admin.ts`).

### `src/design-system` — componentes universais

Componentes visuais reutilizáveis por **todas as páginas**, para manter o
mesmo estilo. Nada de regra de negócio aqui.

```text
design-system/
  index.ts        exports centrais
  layout/         PageHeader, Surface
  forms/          ServerForm, AsyncButton
  feedback/       EmptyState
  branding/       BrandMark
  theme/          theme.ts (createAppTheme) e ThemeRegistry
```

Padrões obrigatórios nas páginas:

- **`PageHeader`** para título + descrição (+ ação opcional, como um botão
  "Novo"). Não criar `<Typography variant="h4/h5">` solto como cabeçalho.
- **`EmptyState`** para listas vazias. Não criar frases soltas.
- **`Surface`** para blocos de conteúdo em destaque (ex.: formulários admin).
- **`ServerForm`** para formulários que chamam server actions com feedback
  de sucesso/erro.
- Tema sempre via `createAppTheme(settings)` em `ThemeRegistry`; cores novas
  entram pelo tema, não por hex hardcoded nas páginas.

### `src/components` — layouts compartilhados

Somente o "esqueleto" das áreas:

- `public/Header.tsx` e `public/Footer.tsx` (recebem `SiteSettings` por props).
- `admin/AdminShell.tsx` (sidebar + topo do painel).

Se um componente visual for usado por duas slices diferentes, ele pertence ao
`design-system`, não a uma delas.

### `src/lib` — infraestrutura transversal

Somente código de infraestrutura compartilhado:

- `supabase/` (browser, server, admin), `stripe/`, `email/`, `storage/`
- `auth/` (roles, authorization), `env.ts`, `audit.ts`
- utilidades puros: `format.ts`, `actions.ts` (`voidAction`), `http.ts`
- `db-types.ts`: tipos das tabelas (candidato a divisão futura por slice)

Não colocar regra de negócio de uma funcionalidade aqui.

## Fluxos importantes

### Pagamentos (Stripe)

1. Formulário (`features/payments/components/TitheForm`) chama
   `/api/payments/create-payment` (avulso) ou `/create-recurring` (setup).
2. A rota valida com Zod, prepara o dizimista e cria o intent na Stripe.
3. O navegador confirma com Payment Element; recorrência cria a assinatura
   via `/api/payments/create-subscription`.
4. **Somente o webhook** (`/api/stripe/webhook` → `payments/webhook.ts`)
   marca pagamentos como `paid`, com idempotência por evento.

### Pix estático

Configurado em `/admin/configuracoes` (slice `settings`). O QR é gerado por
`features/payments/pix.ts` e exibido em `/doacoes`. Não há conciliação
automática: é um QR de valor aberto.

### Cancelamento de recorrência pelo fiel

1. `/doacoes/gerenciar` solicita link por e-mail (`recurring/services`).
2. Token de uso único (hash no banco) cria cookie de sessão curto.
3. Cancelamento chama `payments/services.cancelRecurringAtPeriodEnd` e é
   confirmado pelo webhook da Stripe.

## Banco de dados

- Toda mudança de esquema é uma **migration nova** em `supabase/migrations/`
  (`000NN_descricao.sql`), nunca editar uma migration já aplicada.
- RLS ativa em todas as tabelas; acesso público só ao que é público
  (eventos publicados, períodos publicados, `site_settings`).
- O servidor usa `createAdminClient` (service role) apenas em rotas/actions
  confiáveis; o cliente usa `createClient` com RLS.

## Segurança

- Papéis: `admin` (Pároco na interface) e `assistente`. Verificação sempre
  no servidor (`requireRole`), além do RLS.
- Segredos apenas em variáveis sem `NEXT_PUBLIC_` e nunca versionados
  (ver `.gitignore` e `docs/configuracao.md`).

## Documentação

- `docs/arquitetura.md` (este arquivo): organização do código.
- `docs/configuracao.md`: variáveis de ambiente, modo demo, Pix e migrations.
- `README.md`: visão de produto, licença e autor.
- `.opencode/`: anotações pessoais locais (não versionado).
