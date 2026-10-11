-- name: RejectEvent :exec
INSERT INTO notifications.rejected_events(topic,partition,record_offset,reason) VALUES($1,$2,$3,$4)
ON CONFLICT DO NOTHING;

-- name: AcceptEvent :execrows
INSERT INTO notifications.inbox(id,payload) VALUES($1,$2) ON CONFLICT DO NOTHING;

-- name: CreateDeliveryJobs :exec
INSERT INTO notifications.jobs(event_id,subscription_id)
SELECT sqlc.arg(event_id)::text,id FROM notifications.subscriptions
WHERE user_id=sqlc.arg(user_id) AND expires_at>now() AND session_id<>'';
