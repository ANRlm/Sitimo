-- +goose Up
-- +goose StatementBegin
ALTER TABLE export_jobs
ADD COLUMN IF NOT EXISTS cancel_requested_at timestamptz;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE export_jobs
DROP COLUMN IF EXISTS cancel_requested_at;
-- +goose StatementEnd
