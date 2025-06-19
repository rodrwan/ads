-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS advertisers (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    balance DECIMAL(12, 2) NOT NULL DEFAULT 0.0,
    created_at TIMESTAMP DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS advertisers CASCADE;
-- +goose StatementEnd
