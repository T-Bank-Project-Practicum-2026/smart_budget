-- +goose Up
CREATE TABLE notifications.subscriptions (
 id text PRIMARY KEY,
 user_id text NOT NULL,
 endpoint text NOT NULL UNIQUE,
 p256dh text NOT NULL,
 auth text NOT NULL,
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON notifications.subscriptions(user_id);
CREATE TABLE notifications.inbox (
 id text PRIMARY KEY,
 payload jsonb NOT NULL,
 received_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE notifications.jobs (
 event_id text NOT NULL REFERENCES notifications.inbox(id),
 subscription_id text NOT NULL REFERENCES notifications.subscriptions(id) ON DELETE CASCADE,
 state text NOT NULL DEFAULT 'pending' CHECK (state IN ('pending','sending','delivered','failed')),
 attempts integer NOT NULL DEFAULT 0,
 next_at timestamptz NOT NULL DEFAULT now(),
 lease_token text,
 lease_until timestamptz,
 last_status integer,
 PRIMARY KEY (event_id, subscription_id)
);
CREATE INDEX ON notifications.jobs(next_at) WHERE state IN ('pending','sending');
CREATE TABLE notifications.rejected_events (
 topic text NOT NULL,
 partition integer NOT NULL,
 record_offset bigint NOT NULL,
 PRIMARY KEY(topic, partition, record_offset)
);
GRANT SELECT, INSERT, UPDATE, DELETE ON notifications.subscriptions, notifications.jobs TO budget_notifications;
GRANT SELECT, INSERT ON notifications.inbox, notifications.rejected_events TO budget_notifications;

-- +goose Down
DROP TABLE notifications.jobs, notifications.inbox, notifications.subscriptions, notifications.rejected_events;
