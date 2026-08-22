-- ============================================================
-- Comunidade digitada livremente no formulário de dízimo
-- Permite que o fiel informe o nome da comunidade quando ela
-- não estiver cadastrada (por exemplo, fiéis de outra cidade).
-- ============================================================

alter table public.donors
  add column if not exists community_name text;

alter table public.contribution_payments
  add column if not exists community_name text;