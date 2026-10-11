-- name: DeleteUserSubscription :exec
DELETE FROM notifications.subscriptions WHERE user_id=$1 AND endpoint=$2;

-- name: DeleteSubscription :exec
DELETE FROM notifications.subscriptions WHERE id=$1;

-- name: UpsertSubscription :execrows
INSERT INTO notifications.subscriptions(id,user_id,endpoint,p256dh,auth,session_id,expires_at) VALUES($1,$2,$3,$4,$5,$6,$7)
ON CONFLICT(endpoint) DO UPDATE
SET p256dh=EXCLUDED.p256dh,auth=EXCLUDED.auth,session_id=EXCLUDED.session_id,expires_at=EXCLUDED.expires_at,updated_at=now()
WHERE notifications.subscriptions.user_id=EXCLUDED.user_id;

-- name: DeleteSessionSubscription :exec
DELETE FROM notifications.subscriptions WHERE id=$1 AND session_id=$2;
