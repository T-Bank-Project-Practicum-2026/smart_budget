-- name: LockPushUser :one
SELECT last_push_at FROM app.users WHERE id=$1 FOR UPDATE;

-- name: RecordPushTime :exec
UPDATE app.users SET last_push_at=now() WHERE id=$1;
