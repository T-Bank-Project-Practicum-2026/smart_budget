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

CREATE TABLE app.auth_attempts (
 email_hash text PRIMARY KEY,
 window_start timestamptz NOT NULL DEFAULT now(),
 attempts integer NOT NULL CHECK (attempts > 0)
);

CREATE TABLE app.notifications (
 id text PRIMARY KEY,
 user_id text NOT NULL REFERENCES app.users(id),
 request_key text NOT NULL,
 title text NOT NULL,
 body text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(user_id, request_key)
);
CREATE TABLE app.outbox (
 id text PRIMARY KEY REFERENCES app.notifications(id),
 payload jsonb NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 published_at timestamptz
);
CREATE INDEX ON app.outbox(created_at) WHERE published_at IS NULL;

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

GRANT USAGE ON SCHEMA app TO budget_api, budget_worker;
GRANT USAGE ON SCHEMA notifications TO budget_notifications;
GRANT SELECT, INSERT, UPDATE ON app.users, app.auth_sessions, app.auth_attempts, app.ledger_periods TO budget_api;
GRANT SELECT, INSERT ON app.notifications, app.outbox, app.audit_log TO budget_api;
GRANT SELECT, UPDATE ON app.outbox TO budget_worker;

-- +goose Down
DROP TABLE app.outbox, app.notifications, app.audit_log, app.ledger_periods, app.auth_attempts, app.auth_sessions, app.users;
DROP SCHEMA app;