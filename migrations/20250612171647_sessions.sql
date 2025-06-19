-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS sessions (
    token UUID PRIMARY KEY,
    advertiser_id UUID REFERENCES advertisers(id),
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT now()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd
