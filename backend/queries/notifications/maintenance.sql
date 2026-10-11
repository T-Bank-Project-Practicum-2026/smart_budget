-- name: PruneSubscriptions :execrows
DELETE FROM notifications.subscriptions WHERE id IN (
 SELECT id FROM notifications.subscriptions WHERE expires_at<=now() LIMIT 500
);

-- name: PruneJobs :execrows
DELETE FROM notifications.jobs WHERE (event_id,subscription_id) IN (
 SELECT j.event_id,j.subscription_id FROM notifications.jobs j JOIN notifications.inbox i ON i.id=j.event_id
 WHERE i.received_at<now()-interval '30 days' LIMIT 500
);

-- name: PruneInbox :execrows
DELETE FROM notifications.inbox WHERE id IN (
 SELECT i.id FROM notifications.inbox i WHERE i.received_at<now()-interval '30 days'
 AND NOT EXISTS (SELECT 1 FROM notifications.jobs j WHERE j.event_id=i.id) LIMIT 500
);

-- name: PruneRejectedEvents :execrows
DELETE FROM notifications.rejected_events WHERE (topic,partition,record_offset) IN (
 SELECT topic,partition,record_offset FROM notifications.rejected_events WHERE received_at<now()-interval '30 days' LIMIT 500
);

-- name: DeliveryStatus :one
SELECT count(*) FILTER (WHERE state IN ('pending','sending'))::bigint AS pending,
 count(*) FILTER (WHERE state='failed')::bigint AS failed
FROM notifications.jobs;
