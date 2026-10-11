-- name: EnsureLedgerPeriod :exec
INSERT INTO app.ledger_periods(user_id,month,timezone) VALUES($1,$2,$3) ON CONFLICT DO NOTHING;

-- name: LockLedgerPeriod :one
SELECT timezone FROM app.ledger_periods WHERE user_id=$1 AND month=$2 FOR UPDATE;

-- name: CreateAuditEntry :exec
INSERT INTO app.audit_log(id,user_id,actor_id,entity_type,entity_id,action,old_value,new_value)
VALUES($1,$2,$3,$4,$5,$6,$7,$8);
