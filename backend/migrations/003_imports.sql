-- +goose Up
CREATE TABLE app.imports (
    id text PRIMARY KEY,
    user_id text NOT NULL REFERENCES app.users(id),
    idempotency_key text CHECK (idempotency_key IS NULL OR length(idempotency_key) BETWEEN 16 AND 100),
    file_name text NOT NULL CHECK (btrim(file_name) <> ''),
    file_hash text NOT NULL CHECK (file_hash <> ''),
    status text NOT NULL DEFAULT 'processing' CHECK (status IN ('processing', 'completed', 'failed')),
    total_rows integer NOT NULL DEFAULT 0 CHECK (total_rows >= 0),
    imported_rows integer NOT NULL DEFAULT 0 CHECK (imported_rows >= 0),
    skipped_rows integer NOT NULL DEFAULT 0 CHECK (skipped_rows >= 0),
    duplicate_rows integer NOT NULL DEFAULT 0 CHECK (duplicate_rows >= 0),
    conflict_rows integer NOT NULL DEFAULT 0 CHECK (conflict_rows >= 0),
    failed_rows integer NOT NULL DEFAULT 0 CHECK (failed_rows >= 0),
    errors jsonb NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(errors) = 'array'),
    errors_truncated boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    finished_at timestamptz,
    UNIQUE (user_id, idempotency_key),
    CHECK (finished_at IS NULL OR finished_at >= created_at)
);
CREATE INDEX ON app.imports(user_id, created_at DESC, id DESC);
CREATE INDEX ON app.imports(user_id, file_hash);

CREATE TABLE app.operations (
    id text PRIMARY KEY,
    user_id text NOT NULL REFERENCES app.users(id),
    import_id text NOT NULL REFERENCES app.imports(id),
    external_id text NOT NULL CHECK (external_id <> ''),
    occurred_at timestamptz NOT NULL,
    month date NOT NULL CHECK (extract(day FROM month) = 1),
    amount_kopecks bigint NOT NULL,
    currency text NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    merchant text NOT NULL,
    mcc text NOT NULL DEFAULT '' CHECK (mcc = '' OR mcc ~ '^[0-9]{4}$'),
    row_hash text NOT NULL CHECK (row_hash <> ''),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, external_id)
);
CREATE INDEX ON app.operations(user_id, month, occurred_at DESC, id DESC);
CREATE INDEX ON app.operations(import_id);

GRANT SELECT, INSERT, UPDATE ON app.imports, app.operations TO budget_api;

-- +goose Down
DROP TABLE app.operations, app.imports;
