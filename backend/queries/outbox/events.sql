-- name: LockNextEvent :one
SELECT id,payload,attempts FROM app.outbox WHERE published_at IS NULL AND failed_at IS NULL AND next_at<=now()
ORDER BY created_at LIMIT 1 FOR UPDATE SKIP LOCKED;

-- name: MarkPublished :exec
UPDATE app.outbox SET published_at=now(),last_error=NULL WHERE id=$1;

-- name: QuarantineEvent :exec
UPDATE app.outbox SET failed_at=now(),last_error='invalid_payload',attempts=attempts+1 WHERE id=$1;

-- name: RetryEvent :exec
UPDATE app.outbox SET last_error='broker_unavailable',attempts=attempts+1,next_at=$2 WHERE id=$1;
