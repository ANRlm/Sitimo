-- +goose Up
-- +goose StatementBegin
ALTER TABLE paper_items
ADD COLUMN IF NOT EXISTS blank_lines INTEGER NOT NULL DEFAULT 0
CHECK (blank_lines >= 0 AND blank_lines <= 10);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE paper_items DROP COLUMN IF EXISTS blank_lines;
-- +goose StatementEnd
