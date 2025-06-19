-- +goose Up
-- +goose StatementBegin
CREATE TABLE bids (
    id UUID PRIMARY KEY,
    ad_id UUID REFERENCES ads(id),
    placement_id UUID REFERENCES placements(id),
    bid_price DECIMAL(8, 4) NOT NULL, -- lo que está dispuesto a pagar por 1 impresión
    max_daily_spend DECIMAL(12, 2),
    created_at TIMESTAMP DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS bids CASCADE;
-- +goose StatementEnd
