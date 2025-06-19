-- +goose Up
-- +goose StatementBegin
CREATE TABLE impressions (
    id UUID PRIMARY KEY,
    ad_id UUID REFERENCES ads(id),
    placement_id UUID REFERENCES placements(id),
    auction_id UUID REFERENCES auctions(id),
    user_context JSONB,
    timestamp TIMESTAMP DEFAULT now()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS impressions CASCADE;
-- +goose StatementEnd
