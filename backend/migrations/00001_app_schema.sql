-- +goose Up
CREATE SCHEMA app;

-- +goose Down
DROP SCHEMA app;
