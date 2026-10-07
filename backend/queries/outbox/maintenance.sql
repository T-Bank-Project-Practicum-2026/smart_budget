-- name: PruneSessions :execrows
DELETE FROM app.auth_sessions WHERE id IN (
 SELECT id FROM app.auth_sessions WHERE expires_at<now()-interval '7 days' OR revoked_at<now()-interval '7 days' LIMIT 500
);

-- name: PruneAuthAttempts :execrows
DELETE FROM app.auth_attempts WHERE email_hash IN (
 SELECT email_hash FROM app.auth_attempts WHERE window_start<now()-interval '1 day' LIMIT 500
);

-- name: PrunePublishedEvents :execrows
DELETE FROM app.outbox WHERE id IN (
 SELECT id FROM app.outbox WHERE published_at<now()-interval '30 days' LIMIT 500
);

-- name: OutboxStatus :one
SELECT count(*) FILTER (WHERE published_at IS NULL AND failed_at IS NULL)::bigint AS pending,
 count(*) FILTER (WHERE failed_at IS NOT NULL)::bigint AS quarantined
FROM app.outbox;
