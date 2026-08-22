-- ============================================================
-- Sistema Paróquia - Migração inicial
-- ============================================================

create extension if not exists "pgcrypto";

-- ---------- Helpers de autorização ----------
create or replace function public.set_updated_at()
returns trigger
language plpgsql
as $$
begin
  new.updated_at = now();
  return new;
end;
$$;

create or replace function public.moddatetime()
returns trigger
language plpgsql
as $$
begin
  new.updated_at = now();
  return new;
end;
$$;
create or replace function public.is_staff()
returns boolean
language sql
security definer
set search_path = public
stable
as $$
  select exists (
    select 1
    from public.profiles p
    where p.id = auth.uid()
      and p.is_active = true
      and p.role in ('admin', 'assistente')
  );
$$;

create or replace function public.is_admin()
returns boolean
language sql
security definer
set search_path = public
stable
as $$
  select exists (
    select 1
    from public.profiles p
    where p.id = auth.uid()
      and p.is_active = true
      and p.role = 'admin'
  );
$$;

-- ---------- profiles ----------
create table public.profiles (
  id uuid primary key references auth.users(id) on delete cascade,
  full_name text not null default '',
  email text not null,
  role text not null check (role in ('admin', 'assistente')),
  is_active boolean not null default true,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create trigger profiles_set_updated_at
  before update on public.profiles
  for each row execute procedure moddatetime(updated_at);

alter table public.profiles enable row level security;

create policy "profiles_select_own"
  on public.profiles for select
  using (auth.uid() = id or is_staff());

create policy "profiles_update_own"
  on public.profiles for update
  using (auth.uid() = id)
  with check (auth.uid() = id);

-- Impede que o cliente altere o papel diretamente (deve passar pelo servidor).
create or replace function public.prevent_role_change()
returns trigger
language plpgsql
as $$
begin
  if old.role is distinct from new.role then
    raise exception 'role cannot be changed via client';
  end if;
  return new;
end;
$$;

create trigger profiles_prevent_role_change
  before update on public.profiles
  for each row execute procedure public.prevent_role_change();

-- ---------- communities ----------
create table public.communities (
  id uuid primary key default gen_random_uuid(),
  name text not null,
  description text not null default '',
  city text not null default '',
  is_active boolean not null default true,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

alter table public.communities enable row level security;

create policy "communities_public_read"
  on public.communities for select
  using (is_active = true);

create policy "communities_staff_all"
  on public.communities for all
  using (is_staff())
  with check (is_staff());

-- ---------- donors ----------
create table public.donors (
  id uuid primary key default gen_random_uuid(),
  full_name text not null,
  email text not null,
  phone text,
  city text not null default '',
  neighborhood text not null default '',
  community_id uuid references public.communities(id) on delete set null,
  privacy_consent_at timestamptz,
  is_active boolean not null default true,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index donors_email_idx on public.donors (lower(email));
create index donors_phone_idx on public.donors (phone);
create index donors_community_idx on public.donors (community_id);

alter table public.donors enable row level security;

-- Nenhuma política de leitura pública: dados de dizimistas são restritos.
create policy "donors_staff_all"
  on public.donors for all
  using (is_staff())
  with check (is_staff());

-- ---------- recurring_donations ----------
create table public.recurring_donations (
  id uuid primary key default gen_random_uuid(),
  donor_id uuid not null references public.donors(id) on delete cascade,
  community_id uuid references public.communities(id) on delete set null,
  amount_cents integer not null check (amount_cents > 0),
  currency text not null default 'brl',
  interval text not null default 'month' check (interval = 'month'),
  stripe_customer_id text not null,
  stripe_subscription_id text not null unique,
  status text not null default 'incomplete'
    check (status in ('active', 'past_due', 'cancel_at_period_end', 'canceled', 'incomplete')),
  started_at timestamptz,
  cancel_at_period_end boolean not null default false,
  cancellation_requested_at timestamptz,
  canceled_at timestamptz,
  current_period_end timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index recurring_donations_donor_idx on public.recurring_donations (donor_id);
create index recurring_donations_status_idx on public.recurring_donations (status);

alter table public.recurring_donations enable row level security;

create policy "recurring_donations_staff_all"
  on public.recurring_donations for all
  using (is_staff())
  with check (is_staff());

-- ---------- contribution_payments ----------
create table public.contribution_payments (
  id uuid primary key default gen_random_uuid(),
  donor_id uuid not null references public.donors(id) on delete restrict,
  community_id uuid references public.communities(id) on delete set null,
  recurring_donation_id uuid references public.recurring_donations(id) on delete set null,
  amount_cents integer not null check (amount_cents > 0),
  currency text not null default 'brl',
  payment_type text not null check (payment_type in ('one_time', 'recurring')),
  status text not null default 'pending'
    check (status in ('pending', 'paid', 'failed', 'canceled', 'refunded')),
  stripe_payment_intent_id text,
  stripe_invoice_id text,
  stripe_subscription_id text,
  stripe_event_id text,
  paid_at timestamptz,
  refunded_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create unique index contribution_payments_payment_intent_uidx
  on public.contribution_payments (stripe_payment_intent_id)
  where stripe_payment_intent_id is not null;

create unique index contribution_payments_invoice_uidx
  on public.contribution_payments (stripe_invoice_id)
  where stripe_invoice_id is not null;

create index contribution_payments_donor_idx on public.contribution_payments (donor_id);
create index contribution_payments_community_idx on public.contribution_payments (community_id);
create index contribution_payments_status_idx on public.contribution_payments (status);
create index contribution_payments_paid_at_idx on public.contribution_payments (paid_at);

alter table public.contribution_payments enable row level security;

create policy "contribution_payments_staff_all"
  on public.contribution_payments for all
  using (is_staff())
  with check (is_staff());

-- ---------- events ----------
create table public.events (
  id uuid primary key default gen_random_uuid(),
  title text not null,
  slug text not null unique,
  description text not null default '',
  image_path text,
  starts_at timestamptz not null,
  ends_at timestamptz,
  location text not null default '',
  status text not null default 'draft'
    check (status in ('draft', 'published', 'archived')),
  created_by uuid references public.profiles(id),
  updated_by uuid references public.profiles(id),
  published_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index events_status_idx on public.events (status);
create index events_starts_at_idx on public.events (starts_at desc);

alter table public.events enable row level security;

create policy "events_public_read_published"
  on public.events for select
  using (status = 'published');

create policy "events_staff_all"
  on public.events for all
  using (is_staff())
  with check (is_staff());

-- ---------- contact_messages ----------
create table public.contact_messages (
  id uuid primary key default gen_random_uuid(),
  name text not null,
  email text not null,
  phone text,
  message text not null,
  privacy_consent_at timestamptz not null default now(),
  ip_hash text,
  status text not null default 'new'
    check (status in ('new', 'in_progress', 'resolved', 'archived')),
  assigned_to uuid references public.profiles(id),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index contact_messages_status_idx on public.contact_messages (status);

alter table public.contact_messages enable row level security;

-- Inserção pública (formulário de contato).
create policy "contact_messages_public_insert"
  on public.contact_messages for insert
  with check (true);

create policy "contact_messages_staff_all"
  on public.contact_messages for all
  using (is_staff())
  with check (is_staff());

-- ---------- contact_rate_limit ----------
create table public.contact_rate_limit (
  id uuid primary key default gen_random_uuid(),
  kind text not null default 'contact',
  ip_hash text not null,
  created_at timestamptz not null default now()
);

create index contact_rate_limit_ip_idx on public.contact_rate_limit (kind, ip_hash, created_at);

alter table public.contact_rate_limit enable row level security;

-- Só o servidor insere (via service role).

-- ---------- financial_periods ----------
create table public.financial_periods (
  id uuid primary key default gen_random_uuid(),
  name text not null,
  start_date date not null,
  end_date date not null,
  status text not null default 'draft'
    check (status in ('draft', 'review', 'published', 'archived')),
  created_by uuid references public.profiles(id),
  reviewed_by uuid references public.profiles(id),
  published_by uuid references public.profiles(id),
  published_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index financial_periods_status_idx on public.financial_periods (status);

alter table public.financial_periods enable row level security;

create policy "financial_periods_public_read_published"
  on public.financial_periods for select
  using (status = 'published');

create policy "financial_periods_staff_all"
  on public.financial_periods for all
  using (is_staff())
  with check (is_staff());

-- Só o pároco pode publicar/arquivar um período.
create or replace function public.check_transparency_publish()
returns trigger
language plpgsql
as $$
begin
  if new.status in ('published', 'archived')
     and (old.status is distinct from new.status)
     and not is_admin() then
    raise exception 'apenas o admin pode publicar ou arquivar periodos';
  end if;
  return new;
end;
$$;

create trigger financial_periods_check_publish
  before update on public.financial_periods
  for each row execute procedure public.check_transparency_publish();

-- ---------- financial_entries ----------
create table public.financial_entries (
  id uuid primary key default gen_random_uuid(),
  financial_period_id uuid not null references public.financial_periods(id) on delete cascade,
  type text not null check (type in ('income', 'expense')),
  category text not null,
  description text not null,
  amount_cents integer not null check (amount_cents > 0),
  entry_date date not null,
  source text not null default 'manual' check (source in ('online_contribution', 'manual', 'other')),
  created_by uuid references public.profiles(id),
  updated_by uuid references public.profiles(id),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index financial_entries_period_idx on public.financial_entries (financial_period_id);

alter table public.financial_entries enable row level security;

create policy "financial_entries_public_read_published"
  on public.financial_entries for select
  using (exists (
    select 1 from public.financial_periods fp
    where fp.id = financial_period_id and fp.status = 'published'
  ));

create policy "financial_entries_staff_all"
  on public.financial_entries for all
  using (is_staff())
  with check (is_staff());

-- ---------- transparency_documents ----------
create table public.transparency_documents (
  id uuid primary key default gen_random_uuid(),
  financial_period_id uuid not null references public.financial_periods(id) on delete cascade,
  file_path text not null,
  original_file_name text not null,
  mime_type text not null,
  file_size integer not null,
  created_by uuid references public.profiles(id),
  created_at timestamptz not null default now()
);

alter table public.transparency_documents enable row level security;

create policy "transparency_documents_staff_all"
  on public.transparency_documents for all
  using (is_staff())
  with check (is_staff());

-- Nenhuma leitura pública direta: acesso via rota que valida publicação e gera URL assinada.

-- ---------- stripe_webhook_events ----------
create table public.stripe_webhook_events (
  id uuid primary key default gen_random_uuid(),
  stripe_event_id text not null unique,
  event_type text not null,
  payload_hash text not null,
  processing_status text not null default 'processed'
    check (processing_status in ('processed', 'failed', 'skipped')),
  error_message text,
  created_at timestamptz not null default now()
);

alter table public.stripe_webhook_events enable row level security;

-- Sem políticas: acessado somente via service role.

-- ---------- audit_logs ----------
create table public.audit_logs (
  id uuid primary key default gen_random_uuid(),
  actor_id uuid references public.profiles(id),
  action text not null,
  entity_type text not null,
  entity_id text,
  metadata jsonb,
  created_at timestamptz not null default now()
);

create index audit_logs_entity_idx on public.audit_logs (entity_type, entity_id);

alter table public.audit_logs enable row level security;

create policy "audit_logs_staff_select"
  on public.audit_logs for select
  using (is_staff());

-- ---------- recurring_magic_tokens ----------
create table public.recurring_magic_tokens (
  id uuid primary key default gen_random_uuid(),
  token_hash text not null unique,
  email text not null,
  expires_at timestamptz not null,
  consumed_at timestamptz,
  created_at timestamptz not null default now()
);

create index recurring_magic_tokens_email_idx on public.recurring_magic_tokens (email);

alter table public.recurring_magic_tokens enable row level security;

-- Sem políticas: acessado somente via service role (servidor).

-- ---------- Seeds ----------
insert into public.communities (name, description, city) values
  ('Matriz São Benedito', 'Sede da paróquia', 'Francisco Morato'),
  ('Comunidade São Luís Gonzaga', '', 'Francisco Morato'),
  ('Comunidade Santa Rosa de Lima', '', 'Francisco Morato'),
  ('Comunidade Santa Catarina de Sena', '', 'Francisco Morato')
on conflict do nothing;

-- ---------- Storage ----------
insert into storage.buckets (id, name, public)
values
  ('events', 'events', true),
  ('transparency', 'transparency', false)
on conflict (id) do nothing;

-- Imagens de eventos: leitura pública para qualquer pessoa.
create policy "events_public_read"
  on storage.objects for select
  using (bucket_id = 'events');

-- Upload/remoção de imagens de eventos: apenas equipe.
create policy "events_staff_insert"
  on storage.objects for insert
  with check (bucket_id = 'events' and is_staff());

create policy "events_staff_delete"
  on storage.objects for delete
  using (bucket_id = 'events' and is_staff());

-- Documentos de transparência: bucket privado, acessado somente pelo servidor
-- (URLs assinadas). Sem políticas públicas de leitura.