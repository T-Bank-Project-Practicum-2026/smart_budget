-- name: FindOperationsByExternalIDs :many
SELECT external_id,row_hash FROM app.operations
WHERE user_id=$1 AND external_id = ANY(sqlc.arg(external_ids)::text[]);

-- name: InsertOperation :exec
INSERT INTO app.operations(id,user_id,import_id,external_id,occurred_at,month,amount_kopecks,currency,merchant,mcc,row_hash)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11);

-- name: ListOperations :many
SELECT id,external_id,occurred_at,month,amount_kopecks,currency,merchant,mcc,import_id,created_at
FROM app.operations WHERE user_id=$1 ORDER BY occurred_at DESC, id DESC LIMIT 100;
