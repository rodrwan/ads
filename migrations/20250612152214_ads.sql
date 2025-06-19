-- +goose Up
-- +goose StatementBegin
CREATE TABLE ads (
    id UUID PRIMARY KEY,
    campaign_id UUID REFERENCES ad_campaigns(id),
    title TEXT,
    description TEXT,
    image_url TEXT,
    destination_url TEXT,
    cta_label TEXT,
    targeting_json JSONB, -- condiciones de targeting como país, edad, dispositivo
    status TEXT CHECK (status IN ('active', 'paused')) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ads CASCADE;
-- +goose StatementEnd
