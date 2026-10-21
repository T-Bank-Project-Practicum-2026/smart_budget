-- name: CreateBankConnection :one
INSERT INTO app.bank_connections(id, user_id, provider, external_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetBankConnection :one
SELECT * FROM app.bank_connections WHERE user_id=$1 AND id=$2;

-- name: ListBankConnections :many
SELECT * FROM app.bank_connections WHERE user_id=$1 ORDER BY created_at, id;

-- name: CreateTransaction :one
-- Блокировка месяца, аудит и outbox — в одной транзакции.
-- Дубль или неверный месяц: строк нет; при дубле импорт сверяет данные.
INSERT INTO app.transactions(id, user_id, bank_connection_id, external_id, month,
 occurred_at, kind, amount_kopecks, currency, merchant, mcc, description)
SELECT sqlc.arg(id), p.user_id, sqlc.arg(bank_connection_id), sqlc.arg(external_id), p.month,
 sqlc.arg(occurred_at)::timestamptz, sqlc.arg(kind), sqlc.arg(amount_kopecks),
 sqlc.arg(currency), sqlc.arg(merchant), sqlc.narg(mcc), sqlc.arg(description)
FROM app.ledger_periods p
WHERE p.user_id=sqlc.arg(user_id) AND p.month=sqlc.arg(month)
 AND date_trunc('month', sqlc.arg(occurred_at)::timestamptz AT TIME ZONE p.timezone)::date=p.month
ON CONFLICT (bank_connection_id, external_id) DO NOTHING
RETURNING *;

-- name: GetTransactionByExternalID :one
SELECT * FROM app.transactions WHERE user_id=$1 AND bank_connection_id=$2 AND external_id=$3;

-- name: GetTransaction :one
SELECT * FROM app.transactions WHERE user_id=$1 AND id=$2;

-- name: ListTransactions :many
SELECT * FROM app.transactions
WHERE user_id=$1 AND month=$2
ORDER BY occurred_at DESC, id DESC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: GetMonthlyExpense :one
-- SUM(bigint) is numeric: a month can exceed the int64 range of a single operation.
SELECT COALESCE(sum(amount_kopecks), 0)::text AS amount_kopecks
FROM app.transactions WHERE user_id=$1 AND month=$2 AND kind='expense';
