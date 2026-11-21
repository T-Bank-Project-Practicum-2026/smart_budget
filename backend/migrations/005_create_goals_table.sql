-- +goose Up
-- +goose StatementBegin
CREATE TABLE app.goals (
    id text PRIMARY KEY,
    user_id text NOT NULL REFERENCES app.users(id),
    name text NOT NULL CHECK (btrim(name) <> ''),
    target_kopecks bigint NOT NULL CHECK (target_kopecks > 0 AND target_kopecks <= 100000000),
    current_kopecks bigint NOT NULL DEFAULT 0 CHECK (current_kopecks >= 0 AND current_kopecks <= 100000000),
    deadline timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (current_kopecks <= target_kopecks)
);
-- +goose StatementEnd

CREATE INDEX ON app.goals(user_id, created_at DESC, id DESC);

GRANT SELECT, INSERT, UPDATE, DELETE ON app.goals TO budget_api;

-- +goose Down
DROP TABLE app.goals;
