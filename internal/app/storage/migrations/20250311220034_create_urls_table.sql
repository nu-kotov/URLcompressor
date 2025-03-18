-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS urls (
    short_url      TEXT NOT NULL PRIMARY KEY,
    original_url   TEXT NOT NULL,
    correlation_id TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS urls;
-- +goose StatementEnd
