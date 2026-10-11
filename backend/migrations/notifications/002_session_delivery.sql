-- +goose Up
-- Legacy subscriptions stay inactive until the browser registers an authenticated session.
ALTER TABLE notifications.subscriptions
 ADD COLUMN session_id text NOT NULL DEFAULT '',
 ADD COLUMN expires_at timestamptz NOT NULL DEFAULT now();
CREATE INDEX ON notifications.subscriptions(expires_at);
ALTER TABLE notifications.jobs ADD COLUMN last_error text;
ALTER TABLE notifications.rejected_events
 ADD COLUMN reason text NOT NULL DEFAULT 'invalid_payload',
 ADD COLUMN received_at timestamptz NOT NULL DEFAULT now();
CREATE INDEX ON notifications.inbox(received_at);
CREATE INDEX ON notifications.rejected_events(received_at);
GRANT DELETE ON notifications.inbox, notifications.rejected_events TO budget_notifications;

-- +goose Down
REVOKE DELETE ON notifications.inbox, notifications.rejected_events FROM budget_notifications;
DROP INDEX notifications.rejected_events_received_at_idx;
DROP INDEX notifications.inbox_received_at_idx;
ALTER TABLE notifications.rejected_events DROP COLUMN reason, DROP COLUMN received_at;
ALTER TABLE notifications.jobs DROP COLUMN last_error;
DROP INDEX notifications.subscriptions_expires_at_idx;
ALTER TABLE notifications.subscriptions DROP COLUMN session_id, DROP COLUMN expires_at;
