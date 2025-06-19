-- +goose Up
-- +goose StatementBegin
CREATE TABLE auctions (
    id UUID PRIMARY KEY,
    placement_id UUID REFERENCES placements(id),
    request_context JSONB, -- user info, device, location, etc.
    timestamp TIMESTAMP DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS auctions CASCADE;
-- +goose StatementEnd
