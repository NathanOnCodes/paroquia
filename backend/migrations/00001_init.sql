-- ============================================================================
-- 00001_init.sql
--
-- Schema inicial do backend Go. Espelha o schema legado do Supabase
-- (supabase/migrations/00001_init.sql + 00002 + 00003) com as seguintes
-- mudanças:
--
--   * `auth.users` + `profiles` viram a tabela única `users` (credenciais
--     próprias, sem GoTrue).
--   * Row Level Security é removido: a autorização passa a ser feita na
--     aplicação (middleware de RBAC + escopos nos repositórios).
--   * `is_staff()` / `is_admin()` são reescritos para lerem a tabela `users`;
--     a defesa em profundidade que restou são os triggers de integridade, que
--     identificam o ator a partir da GUC `app.actor_id` definida pela
--     aplicação antes de mutações sensíveis.
--   * Novas tabelas de sessão: `refresh_tokens` e `password_reset_tokens`.
--
-- Os CHECK constraints, UNIQUE e índices parciais são mantidos: são a garantia
-- final contra corrupção de dados, independente da aplicação.
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Aplicação
-- ----------------------------------------------------------------------------

-- Identificador do ator da transação corrente. A aplicação faz
--   SET app.actor_id = '<uuid>';
-- antes das mutações sensíveis, via entgo.io/ent/dialect/sql.WithVar.
-- Quando ausente, os triggers abaixo assumem "sem permissão" (fail-closed).
CREATE OR REPLACE FUNCTION app_actor_id() RETURNS uuid
    LANGUAGE sql STABLE
AS $$ SELECT NULLIF(current_setting('app.actor_id', true), '')::uuid $$;

-- Equivalente direto do `is_admin()` legado: o ator precisa ser um usuário
-- ativo com papel 'admin'. Ler o papel da tabela em vez de transportá-lo no
-- GUC garante que uma revogação de acesso valha imediatamente.
CREATE OR REPLACE FUNCTION is_admin() RETURNS boolean
    LANGUAGE sql STABLE SECURITY DEFINER
    SET search_path = public
AS $$
    SELECT EXISTS (
        SELECT 1 FROM users
        WHERE id = app_actor_id()
          AND role = 'admin'
          AND is_active
    )
$$;

CREATE OR REPLACE FUNCTION is_staff() RETURNS boolean
    LANGUAGE sql STABLE SECURITY DEFINER
    SET search_path = public
AS $$
    SELECT EXISTS (
        SELECT 1 FROM users
        WHERE id = app_actor_id()
          AND role IN ('admin', 'assistente')
          AND is_active
    )
$$;

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger
    LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;

-- Bloqueia trocas de papel que não venham de um ator administrador.
CREATE OR REPLACE FUNCTION prevent_role_change() RETURNS trigger
    LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.role IS DISTINCT FROM NEW.role AND NOT is_admin() THEN
        RAISE EXCEPTION 'role cannot be changed without admin actor';
    END IF;
    RETURN NEW;
END;
$$;

-- Apenas um admin pode publicar ou arquivar um período de transparência.
CREATE OR REPLACE FUNCTION check_transparency_publish() RETURNS trigger
    LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.status IN ('published', 'archived')
       AND OLD.status IS DISTINCT FROM NEW.status
       AND NOT is_admin() THEN
        RAISE EXCEPTION 'apenas o admin pode publicar ou arquivar periodos';
    END IF;
    RETURN NEW;
END;
$$;

-- ----------------------------------------------------------------------------
-- users — conta administrativa (substitui auth.users + profiles)
-- ----------------------------------------------------------------------------

CREATE TABLE users (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email                text NOT NULL UNIQUE,
    password_hash        text NOT NULL,
    full_name            text NOT NULL,
    role                 text NOT NULL DEFAULT 'assistente'
                             CHECK (role IN ('admin', 'assistente')),
    is_active            boolean NOT NULL DEFAULT true,
    must_change_password boolean NOT NULL DEFAULT false,
    password_changed_at  timestamptz,
    last_login_at        timestamptz,
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER users_set_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER users_prevent_role_change BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION prevent_role_change();

-- ----------------------------------------------------------------------------
-- refresh_tokens — sessão rotativa (apenas hash SHA-256 é persistido)
-- ----------------------------------------------------------------------------

CREATE TABLE refresh_tokens (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash text NOT NULL UNIQUE,
    user_id    uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    family_id  uuid NOT NULL,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    reused_at  timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX refresh_tokens_user_idx ON refresh_tokens (user_id);
CREATE INDEX refresh_tokens_family_idx ON refresh_tokens (family_id);

-- ----------------------------------------------------------------------------
-- password_reset_tokens — recuperação de acesso (uso único)
-- ----------------------------------------------------------------------------

CREATE TABLE password_reset_tokens (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash  text NOT NULL UNIQUE,
    user_id     uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    expires_at  timestamptz NOT NULL,
    consumed_at timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX password_reset_tokens_user_idx ON password_reset_tokens (user_id);

-- ----------------------------------------------------------------------------
-- communities
-- ----------------------------------------------------------------------------

CREATE TABLE communities (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text NOT NULL,
    description text NOT NULL DEFAULT '',
    city        text NOT NULL DEFAULT '',
    is_active   boolean NOT NULL DEFAULT true,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER communities_set_updated_at BEFORE UPDATE ON communities
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ----------------------------------------------------------------------------
-- donors
--
-- `community_id` OU `community_name` (comunidade livre de outra paróquia).
-- Quando `community_id` está preenchido, `community_name` fica nulo.
-- ----------------------------------------------------------------------------

CREATE TABLE donors (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name           text NOT NULL,
    email               text NOT NULL,
    phone               text,
    city                text NOT NULL DEFAULT '',
    neighborhood        text NOT NULL DEFAULT '',
    community_id        uuid REFERENCES communities (id) ON DELETE SET NULL,
    community_name      text,
    privacy_consent_at  timestamptz,
    is_active           boolean NOT NULL DEFAULT true,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT donors_community_xor_name
        CHECK (community_id IS NULL OR community_name IS NULL)
);

CREATE INDEX donors_email_idx ON donors (lower(email));
CREATE INDEX donors_phone_idx ON donors (phone);
CREATE INDEX donors_community_idx ON donors (community_id);

CREATE TRIGGER donors_set_updated_at BEFORE UPDATE ON donors
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ----------------------------------------------------------------------------
-- recurring_donations — espelho local de uma assinatura Stripe
-- ----------------------------------------------------------------------------

CREATE TABLE recurring_donations (
    id                        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    donor_id                  uuid NOT NULL REFERENCES donors (id) ON DELETE CASCADE,
    community_id              uuid REFERENCES communities (id) ON DELETE SET NULL,
    amount_cents              integer NOT NULL CHECK (amount_cents > 0),
    currency                  text NOT NULL DEFAULT 'brl',
    "interval"                text NOT NULL DEFAULT 'month' CHECK ("interval" = 'month'),
    stripe_customer_id        text NOT NULL,
    stripe_subscription_id    text NOT NULL UNIQUE,
    status                    text NOT NULL DEFAULT 'incomplete'
                                  CHECK (status IN ('incomplete', 'active', 'past_due',
                                                    'cancel_at_period_end', 'canceled')),
    started_at                timestamptz,
    cancel_at_period_end      boolean NOT NULL DEFAULT false,
    cancellation_requested_at timestamptz,
    canceled_at               timestamptz,
    current_period_end        timestamptz,
    created_at                timestamptz NOT NULL DEFAULT now(),
    updated_at                timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX recurring_donations_donor_idx ON recurring_donations (donor_id);
CREATE INDEX recurring_donations_status_idx ON recurring_donations (status);
CREATE INDEX recurring_donations_customer_idx ON recurring_donations (stripe_customer_id);

CREATE TRIGGER recurring_donations_set_updated_at BEFORE UPDATE ON recurring_donations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ----------------------------------------------------------------------------
-- contribution_payments — livro-razão das contribuições
--
-- As linhas são criadas pelo webhook do Stripe, nunca pelo handler público.
-- Os índices únicos parciais abaixo são a garantia de idempotência real.
-- ----------------------------------------------------------------------------

CREATE TABLE contribution_payments (
    id                       uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    donor_id                 uuid NOT NULL REFERENCES donors (id) ON DELETE RESTRICT,
    community_id             uuid REFERENCES communities (id) ON DELETE SET NULL,
    recurring_donation_id    uuid REFERENCES recurring_donations (id) ON DELETE SET NULL,
    community_name           text,
    amount_cents             integer NOT NULL CHECK (amount_cents > 0),
    currency                 text NOT NULL DEFAULT 'brl',
    payment_type             text NOT NULL CHECK (payment_type IN ('one_time', 'recurring')),
    status                   text NOT NULL DEFAULT 'pending'
                                 CHECK (status IN ('pending', 'paid', 'failed', 'canceled', 'refunded')),
    stripe_payment_intent_id text,
    stripe_invoice_id        text,
    stripe_subscription_id   text,
    stripe_event_id          text,
    paid_at                  timestamptz,
    refunded_at              timestamptz,
    created_at               timestamptz NOT NULL DEFAULT now(),
    updated_at               timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX contribution_payments_payment_intent_uidx
    ON contribution_payments (stripe_payment_intent_id)
    WHERE stripe_payment_intent_id IS NOT NULL;

CREATE UNIQUE INDEX contribution_payments_invoice_uidx
    ON contribution_payments (stripe_invoice_id)
    WHERE stripe_invoice_id IS NOT NULL;

CREATE INDEX contribution_payments_donor_idx ON contribution_payments (donor_id);
CREATE INDEX contribution_payments_community_idx ON contribution_payments (community_id);
CREATE INDEX contribution_payments_status_idx ON contribution_payments (status);
CREATE INDEX contribution_payments_paid_at_idx ON contribution_payments (paid_at);

CREATE TRIGGER contribution_payments_set_updated_at BEFORE UPDATE ON contribution_payments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ----------------------------------------------------------------------------
-- events
-- ----------------------------------------------------------------------------

CREATE TABLE events (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    title        text NOT NULL,
    slug         text NOT NULL UNIQUE,
    description  text NOT NULL DEFAULT '',
    image_path   text,
    starts_at    timestamptz NOT NULL,
    ends_at      timestamptz,
    location     text NOT NULL DEFAULT '',
    status       text NOT NULL DEFAULT 'draft'
                     CHECK (status IN ('draft', 'published', 'archived')),
    created_by   uuid REFERENCES users (id),
    updated_by   uuid REFERENCES users (id),
    published_at timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX events_status_idx ON events (status);
CREATE INDEX events_starts_at_idx ON events (starts_at DESC);

CREATE TRIGGER events_set_updated_at BEFORE UPDATE ON events
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ----------------------------------------------------------------------------
-- contact_messages
-- ----------------------------------------------------------------------------

CREATE TABLE contact_messages (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name               text NOT NULL,
    email              text NOT NULL,
    phone              text,
    message            text NOT NULL,
    privacy_consent_at timestamptz NOT NULL DEFAULT now(),
    ip_hash            text,
    status             text NOT NULL DEFAULT 'new'
                           CHECK (status IN ('new', 'in_progress', 'resolved', 'archived')),
    assigned_to        uuid REFERENCES users (id),
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX contact_messages_status_idx ON contact_messages (status);
CREATE INDEX contact_messages_assigned_idx ON contact_messages (assigned_to);

CREATE TRIGGER contact_messages_set_updated_at BEFORE UPDATE ON contact_messages
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ----------------------------------------------------------------------------
-- contact_rate_limit — janela fixa do limitador de taxa
-- ----------------------------------------------------------------------------

CREATE TABLE contact_rate_limit (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    kind       text NOT NULL DEFAULT 'contact',
    ip_hash    text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX contact_rate_limit_ip_idx ON contact_rate_limit (kind, ip_hash, created_at);

-- ----------------------------------------------------------------------------
-- financial_periods
-- ----------------------------------------------------------------------------

CREATE TABLE financial_periods (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name         text NOT NULL,
    start_date   date NOT NULL,
    end_date     date NOT NULL,
    status       text NOT NULL DEFAULT 'draft'
                     CHECK (status IN ('draft', 'review', 'published', 'archived')),
    created_by   uuid REFERENCES users (id),
    reviewed_by  uuid REFERENCES users (id),
    published_by uuid REFERENCES users (id),
    published_at timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT financial_periods_date_order CHECK (start_date <= end_date)
);

CREATE INDEX financial_periods_status_idx ON financial_periods (status);

CREATE TRIGGER financial_periods_set_updated_at BEFORE UPDATE ON financial_periods
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER financial_periods_check_publish BEFORE UPDATE ON financial_periods
    FOR EACH ROW EXECUTE FUNCTION check_transparency_publish();

-- ----------------------------------------------------------------------------
-- financial_entries
-- ----------------------------------------------------------------------------

CREATE TABLE financial_entries (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    financial_period_id uuid NOT NULL REFERENCES financial_periods (id) ON DELETE CASCADE,
    type                text NOT NULL CHECK (type IN ('income', 'expense')),
    category            text NOT NULL,
    description         text NOT NULL,
    amount_cents        integer NOT NULL CHECK (amount_cents > 0 AND amount_cents <= 1000000000),
    entry_date          date NOT NULL,
    source              text NOT NULL DEFAULT 'manual'
                            CHECK (source IN ('online_contribution', 'manual', 'other')),
    created_by          uuid REFERENCES users (id),
    updated_by          uuid REFERENCES users (id),
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX financial_entries_period_idx ON financial_entries (financial_period_id);

CREATE TRIGGER financial_entries_set_updated_at BEFORE UPDATE ON financial_entries
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ----------------------------------------------------------------------------
-- transparency_documents — anexos de um período (DRE, recibos)
-- ----------------------------------------------------------------------------

CREATE TABLE transparency_documents (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    financial_period_id uuid NOT NULL REFERENCES financial_periods (id) ON DELETE CASCADE,
    file_path           text NOT NULL,
    original_file_name  text NOT NULL,
    mime_type           text NOT NULL,
    file_size           integer NOT NULL CHECK (file_size > 0),
    created_by          uuid REFERENCES users (id),
    created_at          timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX transparency_documents_period_idx
    ON transparency_documents (financial_period_id, created_at DESC);

-- ----------------------------------------------------------------------------
-- stripe_webhook_events — livro de idempotência do webhook
-- ----------------------------------------------------------------------------

CREATE TABLE stripe_webhook_events (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    stripe_event_id   text NOT NULL UNIQUE,
    event_type        text NOT NULL,
    payload_hash      text NOT NULL,
    processing_status text NOT NULL DEFAULT 'processed'
                          CHECK (processing_status IN ('processed', 'failed', 'skipped')),
    error_message     text,
    created_at        timestamptz NOT NULL DEFAULT now()
);

-- ----------------------------------------------------------------------------
-- audit_logs
-- ----------------------------------------------------------------------------

CREATE TABLE audit_logs (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id    uuid REFERENCES users (id),
    action      text NOT NULL,
    entity_type text NOT NULL,
    entity_id   text,
    metadata    jsonb,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX audit_logs_entity_idx ON audit_logs (entity_type, entity_id);
CREATE INDEX audit_logs_actor_idx ON audit_logs (actor_id);

-- ----------------------------------------------------------------------------
-- recurring_magic_tokens — link mágico do doador (uso único)
-- ----------------------------------------------------------------------------

CREATE TABLE recurring_magic_tokens (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash  text NOT NULL UNIQUE,
    email       text NOT NULL,
    expires_at  timestamptz NOT NULL,
    consumed_at timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX recurring_magic_tokens_email_idx ON recurring_magic_tokens (email);

-- ----------------------------------------------------------------------------
-- site_settings — linha única (id = 1)
-- ----------------------------------------------------------------------------

CREATE TABLE site_settings (
    id                 integer PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    site_name          text NOT NULL DEFAULT 'Paroquia',
    site_title         text NOT NULL DEFAULT 'Paroquia',
    city               text NOT NULL DEFAULT '',
    description        text NOT NULL DEFAULT 'Comunidade, fe e acolhimento.',
    secondary_color    text NOT NULL DEFAULT '#1D4E89',
    pix_key            text,
    pix_receiver_name  text,
    pix_receiver_city  text,
    home_title         text,
    home_description   text,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER site_settings_set_updated_at BEFORE UPDATE ON site_settings
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
