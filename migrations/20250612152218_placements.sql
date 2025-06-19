-- +goose Up
-- +goose StatementBegin
CREATE TABLE placements (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    slot_size TEXT, -- ej: "300x250", "fullscreen"
    created_at TIMESTAMP DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS placements CASCADE;
-- +goose StatementEnd
