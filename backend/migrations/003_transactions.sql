-- +goose Up
CREATE TABLE app.bank_connections (
 id text PRIMARY KEY,
 user_id text NOT NULL REFERENCES app.users(id),
 provider text NOT NULL CHECK (length(btrim(provider)) > 0),
 external_id text NOT NULL CHECK (length(btrim(external_id)) > 0),
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE (user_id, id),
 UNIQUE (user_id, provider, external_id)
);

CREATE TABLE app.transactions (
 id text PRIMARY KEY,
 user_id text NOT NULL REFERENCES app.users(id),
 bank_connection_id text NOT NULL,
 external_id text NOT NULL CHECK (length(btrim(external_id)) > 0),
 month date NOT NULL,
 occurred_at timestamptz NOT NULL,
 kind text NOT NULL CHECK (kind IN ('expense', 'income', 'transfer')),
 amount_kopecks bigint NOT NULL CHECK (amount_kopecks > 0),
 currency text NOT NULL DEFAULT 'RUB' CHECK (currency = 'RUB'),
 merchant text NOT NULL DEFAULT '',
 mcc text CHECK (mcc ~ '^[0-9]{4}$'),
 description text NOT NULL DEFAULT '',
 created_at timestamptz NOT NULL DEFAULT now(),
 FOREIGN KEY (user_id, bank_connection_id) REFERENCES app.bank_connections(user_id, id),
 FOREIGN KEY (user_id, month) REFERENCES app.ledger_periods(user_id, month),
 UNIQUE (bank_connection_id, external_id)
);
CREATE INDEX transactions_user_month_time_idx ON app.transactions(user_id, month, occurred_at DESC, id DESC);
GRANT SELECT, INSERT ON app.bank_connections, app.transactions TO budget_api;

-- +goose Down
DROP TABLE app.transactions;
DROP TABLE app.bank_connections;
