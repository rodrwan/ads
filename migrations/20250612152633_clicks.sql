-- +goose Up
-- +goose StatementBegin
CREATE TABLE clicks (
    id UUID PRIMARY KEY,
    impression_id UUID REFERENCES impressions(id),
    clicked_at TIMESTAMP DEFAULT now()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS clicks CASCADE;
-- +goose StatementEnd
