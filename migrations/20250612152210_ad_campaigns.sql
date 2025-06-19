-- +goose Up
-- +goose StatementBegin
CREATE TABLE ad_campaigns (
    id UUID PRIMARY KEY,
    advertiser_id UUID REFERENCES advertisers(id),
    name TEXT NOT NULL,
    status TEXT CHECK (status IN ('active', 'paused', 'ended')) DEFAULT 'active',
    budget DECIMAL(12, 2),
    daily_budget DECIMAL(12, 2),
    start_date DATE,
    end_date DATE,
    created_at TIMESTAMP DEFAULT now()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ad_campaigns CASCADE;
-- +goose StatementEnd
