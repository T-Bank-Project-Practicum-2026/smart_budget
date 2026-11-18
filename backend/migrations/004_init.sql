-- +goose Up
CREATE SCHEMA app;
CREATE SCHEMA IF NOT EXISTS notifications;

-- +goose StatementBegin
DO $$ BEGIN
 IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname='budget_api') THEN CREATE ROLE budget_api NOLOGIN; END IF;
 IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname='budget_worker') THEN CREATE ROLE budget_worker NOLOGIN; END IF;
 IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname='budget_notifications') THEN CREATE ROLE budget_notifications NOLOGIN; END IF;
END $$;
-- +goose StatementEnd

CREATE TABLE app.users (
 id text PRIMARY KEY,
 email text NOT NULL UNIQUE,
 password_hash text NOT NULL,
 timezone text NOT NULL DEFAULT 'UTC',
 created_at timestamptz NOT NULL DEFAULT now(),
 last_push_at timestamptz,
 CHECK (email = lower(btrim(email)) AND length(email) BETWEEN 3 AND 254)
);

CREATE TABLE app.auth_sessions (
 id text PRIMARY KEY,
 user_id text NOT NULL REFERENCES app.users(id),
 refresh_hash text NOT NULL UNIQUE,
 created_at timestamptz NOT NULL DEFAULT now(),
 expires_at timestamptz NOT NULL,
 revoked_at timestamptz,
 CHECK (expires_at > created_at)
);
CREATE INDEX ON app.auth_sessions(user_id);
CREATE INDEX ON app.auth_sessions(expires_at);
CREATE INDEX ON app.auth_sessions(revoked_at) WHERE revoked_at IS NOT NULL;

CREATE TABLE app.auth_attempts (
 email_hash text PRIMARY KEY,
 window_start timestamptz NOT NULL DEFAULT now(),
 attempts integer NOT NULL CHECK (attempts > 0)
);
CREATE INDEX ON app.auth_attempts(window_start);

CREATE TABLE app.notifications (
 id text PRIMARY KEY,
 user_id text NOT NULL REFERENCES app.users(id),
 request_key text NOT NULL,
 title text NOT NULL,
 body text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(user_id, request_key)
);
CREATE INDEX ON app.notifications(user_id, created_at DESC, id DESC);
CREATE TABLE app.outbox (
 id text PRIMARY KEY REFERENCES app.notifications(id),
 payload jsonb NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 published_at timestamptz,
 failed_at timestamptz,
 last_error text,
 attempts integer NOT NULL DEFAULT 0,
 next_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX outbox_pending_idx ON app.outbox(next_at, created_at) WHERE published_at IS NULL AND failed_at IS NULL;
CREATE INDEX ON app.outbox(published_at) WHERE published_at IS NOT NULL;

CREATE TABLE app.ledger_periods (
 user_id text NOT NULL REFERENCES app.users(id),
 month date NOT NULL CHECK (extract(day FROM month) = 1),
 timezone text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY (user_id, month)
);
CREATE TABLE app.audit_log (
 id text PRIMARY KEY,
 user_id text NOT NULL REFERENCES app.users(id),
 actor_id text NOT NULL REFERENCES app.users(id),
 entity_type text NOT NULL,
 entity_id text NOT NULL,
 action text NOT NULL,
 old_value jsonb,
 new_value jsonb,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON app.audit_log(user_id, created_at DESC, id);

CREATE TABLE app.imports (
    id text PRIMARY KEY,
    user_id text NOT NULL REFERENCES app.users(id),
    idempotency_key text CHECK (idempotency_key IS NULL OR length(idempotency_key) BETWEEN 16 AND 100),
    file_name text NOT NULL CHECK (btrim(file_name) <> ''),
    file_hash text NOT NULL CHECK (file_hash <> ''),
    status text NOT NULL DEFAULT 'processing' CHECK (status IN ('processing', 'completed', 'failed')),
    total_rows integer NOT NULL DEFAULT 0 CHECK (total_rows >= 0),
    imported_rows integer NOT NULL DEFAULT 0 CHECK (imported_rows >= 0),
    skipped_rows integer NOT NULL DEFAULT 0 CHECK (skipped_rows >= 0),
    duplicate_rows integer NOT NULL DEFAULT 0 CHECK (duplicate_rows >= 0),
    conflict_rows integer NOT NULL DEFAULT 0 CHECK (conflict_rows >= 0),
    failed_rows integer NOT NULL DEFAULT 0 CHECK (failed_rows >= 0),
    errors jsonb NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(errors) = 'array'),
    errors_truncated boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    finished_at timestamptz,
    UNIQUE (user_id, idempotency_key),
    CHECK (finished_at IS NULL OR finished_at >= created_at)
);
CREATE INDEX ON app.imports(user_id, created_at DESC, id DESC);
CREATE INDEX ON app.imports(user_id, file_hash);

CREATE TABLE app.operations (
    id text PRIMARY KEY,
    user_id text NOT NULL REFERENCES app.users(id),
    import_id text NOT NULL REFERENCES app.imports(id),
    external_id text NOT NULL CHECK (external_id <> ''),
    occurred_at timestamptz NOT NULL,
    month date NOT NULL CHECK (extract(day FROM month) = 1),
    amount_kopecks bigint NOT NULL,
    currency text NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    merchant text NOT NULL,
    mcc text NOT NULL DEFAULT '' CHECK (mcc = '' OR mcc ~ '^[0-9]{4}$'),
    row_hash text NOT NULL CHECK (row_hash <> ''),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, external_id)
);
CREATE INDEX ON app.operations(user_id, month, occurred_at DESC, id DESC);
CREATE INDEX ON app.operations(import_id);

GRANT USAGE ON SCHEMA app TO budget_api, budget_worker;
GRANT USAGE ON SCHEMA notifications TO budget_notifications;
GRANT SELECT, INSERT, UPDATE ON app.users, app.auth_sessions, app.auth_attempts, app.ledger_periods, app.imports, app.operations TO budget_api;
GRANT SELECT, INSERT ON app.notifications, app.outbox, app.audit_log TO budget_api;
GRANT SELECT, UPDATE, DELETE ON app.outbox TO budget_worker;
GRANT SELECT, DELETE ON app.auth_sessions, app.auth_attempts TO budget_worker;

-- +goose Down
DROP TABLE app.operations, app.imports;
DROP TABLE app.outbox, app.notifications, app.audit_log, app.ledger_periods, app.auth_attempts, app.auth_sessions, app.users;
DROP SCHEMA app;
