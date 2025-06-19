-- +goose Up
-- +goose StatementBegin
CREATE TABLE auction_bids (
    id UUID PRIMARY KEY,
    auction_id UUID REFERENCES auctions(id),
    bid_id UUID REFERENCES bids(id),
    price DECIMAL(8, 4),
    score NUMERIC, -- puede incluir CTR estimado, relevancia, etc.
    is_winner BOOLEAN DEFAULT FALSE
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS auction_bids CASCADE;
-- +goose StatementEnd
