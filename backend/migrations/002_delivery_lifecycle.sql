-- +goose Up
ALTER TABLE app.outbox
 ADD COLUMN failed_at timestamptz,
 ADD COLUMN last_error text,
 ADD COLUMN attempts integer NOT NULL DEFAULT 0,
 ADD COLUMN next_at timestamptz NOT NULL DEFAULT now();
DROP INDEX app.outbox_created_at_idx;
CREATE INDEX outbox_pending_idx ON app.outbox(next_at, created_at) WHERE published_at IS NULL AND failed_at IS NULL;
CREATE INDEX ON app.outbox(published_at) WHERE published_at IS NOT NULL;
CREATE INDEX ON app.auth_sessions(expires_at);
CREATE INDEX ON app.auth_sessions(revoked_at) WHERE revoked_at IS NOT NULL;
CREATE INDEX ON app.auth_attempts(window_start);
CREATE INDEX ON app.notifications(user_id, created_at DESC, id DESC);
GRANT SELECT, DELETE ON app.auth_sessions, app.auth_attempts TO budget_worker;
GRANT DELETE ON app.outbox TO budget_worker;

-- +goose Down
REVOKE SELECT, DELETE ON app.auth_sessions, app.auth_attempts FROM budget_worker;
REVOKE DELETE ON app.outbox FROM budget_worker;
DROP INDEX app.notifications_user_id_created_at_id_idx;
DROP INDEX app.auth_attempts_window_start_idx;
DROP INDEX app.auth_sessions_revoked_at_idx;
DROP INDEX app.auth_sessions_expires_at_idx;
DROP INDEX app.outbox_published_at_idx;
DROP INDEX app.outbox_pending_idx;
ALTER TABLE app.outbox DROP COLUMN failed_at, DROP COLUMN last_error, DROP COLUMN attempts, DROP COLUMN next_at;
CREATE INDEX ON app.outbox(created_at) WHERE published_at IS NULL;
