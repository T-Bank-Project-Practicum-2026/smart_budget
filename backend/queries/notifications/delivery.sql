-- name: ClaimDelivery :one
WITH candidate AS (
  SELECT event_id,subscription_id FROM notifications.jobs
  WHERE (state='pending' AND next_at<=now()) OR (state='sending' AND lease_until<now())
  ORDER BY next_at LIMIT 1 FOR UPDATE SKIP LOCKED
)
UPDATE notifications.jobs j
SET state='sending',attempts=j.attempts+1,lease_token=sqlc.arg(lease_token)::text,lease_until=now()+interval '60 seconds'
FROM candidate c WHERE j.event_id=c.event_id AND j.subscription_id=c.subscription_id
RETURNING j.event_id,j.subscription_id,j.attempts;

-- name: GetDelivery :one
SELECT i.payload,s.endpoint,s.p256dh,s.auth,s.user_id,s.session_id,s.expires_at FROM notifications.inbox i, notifications.subscriptions s
WHERE i.id=sqlc.arg(event_id) AND s.id=sqlc.arg(subscription_id);

-- name: CompleteDelivery :exec
UPDATE notifications.jobs
SET state=sqlc.arg(state),last_status=sqlc.arg(last_status)::integer,last_error=sqlc.arg(last_error)::text,next_at=sqlc.arg(next_at),lease_token=NULL,lease_until=NULL
WHERE event_id=sqlc.arg(event_id) AND subscription_id=sqlc.arg(subscription_id) AND lease_token=sqlc.arg(lease_token)::text;
