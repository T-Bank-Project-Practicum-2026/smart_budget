-- name: FindNotification :one
SELECT id FROM app.notifications WHERE user_id=$1 AND request_key=$2;

-- name: CreateNotification :exec
INSERT INTO app.notifications(id,user_id,request_key,title,body) VALUES($1,$2,$3,$4,$5);

-- name: EnqueueNotification :exec
INSERT INTO app.outbox(id,payload) VALUES($1,$2);

-- name: ListNotifications :many
SELECT id,title,body,created_at FROM app.notifications
WHERE user_id=$1 ORDER BY created_at DESC LIMIT 20;
