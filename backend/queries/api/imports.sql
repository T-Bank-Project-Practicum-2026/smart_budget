-- name: CreateImport :exec
INSERT INTO app.imports(id,user_id,idempotency_key,file_name,file_hash)
VALUES($1,$2,$3,$4,$5);

-- name: FindImportIDByIdempotencyKey :one
SELECT id FROM app.imports WHERE user_id=$1 AND idempotency_key=$2;

-- name: GetImport :one
SELECT id,user_id,idempotency_key,file_name,file_hash,status,total_rows,imported_rows,skipped_rows,duplicate_rows,conflict_rows,failed_rows,errors,errors_truncated,created_at,finished_at
FROM app.imports WHERE id=$1 AND user_id=$2;

-- name: ListImports :many
SELECT id,user_id,idempotency_key,file_name,file_hash,status,total_rows,imported_rows,skipped_rows,duplicate_rows,conflict_rows,failed_rows,errors,errors_truncated,created_at,finished_at
FROM app.imports WHERE user_id=$1 ORDER BY created_at DESC, id DESC LIMIT 50;

-- name: CompleteImport :exec
UPDATE app.imports
SET status='completed',total_rows=$1,imported_rows=$2,skipped_rows=$3,duplicate_rows=$4,conflict_rows=$5,failed_rows=$6,errors=$7,errors_truncated=$8,finished_at=now()
WHERE id=$9 AND user_id=$10 AND status='processing';

-- name: GetUserTimezone :one
SELECT timezone FROM app.users WHERE id=$1;
