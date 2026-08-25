-- ============================================================
-- Configuracao publica white label
-- ============================================================

create table public.site_settings (
  id integer primary key default 1 check (id = 1),
  site_name text not null default 'Paroquia',
  site_title text not null default 'Paroquia',
  city text not null default '',
  description text not null default 'Comunidade, fe e acolhimento.',
  secondary_color text not null default '#1D4E89',
  pix_key text,
  pix_receiver_name text,
  pix_receiver_city text,
  home_title text,
  home_description text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create trigger site_settings_set_updated_at
  before update on public.site_settings
  for each row execute procedure moddatetime(updated_at);

alter table public.site_settings enable row level security;

create policy "site_settings_public_read"
  on public.site_settings for select
  using (true);

create policy "site_settings_admin_write"
  on public.site_settings for all
  using (is_admin())
  with check (is_admin());

insert into public.site_settings (id)
values (1)
on conflict (id) do nothing;
