-- name: CreateUser :one
INSERT INTO app.users(id,email,password_hash,timezone) VALUES($1,$2,$3,$4)
RETURNING id,email,timezone,created_at;

-- name: FindUserByEmail :one
SELECT id,email,password_hash,timezone,created_at FROM app.users WHERE email=$1;

-- name: GetSessionProfile :one
SELECT u.id,u.email,u.timezone,u.created_at FROM app.users u
JOIN app.auth_sessions s ON s.user_id=u.id
WHERE s.id=$1 AND s.user_id=$2 AND s.revoked_at IS NULL AND s.expires_at>now();

-- name: CreateAuthSession :exec
INSERT INTO app.auth_sessions(id,user_id,refresh_hash,expires_at) VALUES($1,$2,$3,$4);

-- name: GetAuthSessionUser :one
SELECT user_id FROM app.auth_sessions
WHERE id=$1 AND user_id=$2 AND revoked_at IS NULL AND expires_at>now();

-- name: GetActiveSession :one
SELECT id,user_id,expires_at FROM app.auth_sessions
WHERE id=$1 AND user_id=$2 AND revoked_at IS NULL AND expires_at>now();

-- name: RotateRefreshToken :one
UPDATE app.auth_sessions SET refresh_hash=sqlc.arg(new_hash)
WHERE refresh_hash=sqlc.arg(old_hash) AND revoked_at IS NULL AND expires_at>now()
RETURNING id,user_id,expires_at;

-- name: RevokeAuthSession :exec
UPDATE app.auth_sessions SET revoked_at=now() WHERE id=$1 AND user_id=$2 AND revoked_at IS NULL;

-- name: RevokeRefreshToken :exec
UPDATE app.auth_sessions SET revoked_at=now() WHERE refresh_hash=$1 AND revoked_at IS NULL;

-- name: RefreshOwner :one
SELECT user_id FROM app.auth_sessions WHERE refresh_hash=$1;

-- name: CountAuthAttempt :one
INSERT INTO app.auth_attempts(email_hash,attempts) VALUES($1,1)
ON CONFLICT (email_hash) DO UPDATE SET
 attempts=CASE WHEN app.auth_attempts.window_start<=now()-interval '15 minutes' THEN 1 ELSE least(app.auth_attempts.attempts+1,11) END,
 window_start=CASE WHEN app.auth_attempts.window_start<=now()-interval '15 minutes' THEN now() ELSE app.auth_attempts.window_start END
RETURNING attempts;
