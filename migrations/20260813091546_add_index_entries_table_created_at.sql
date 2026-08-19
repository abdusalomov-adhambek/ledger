-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE INDEX IF NOT EXISTS idx_entries_account_created ON entries (account_id, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP INDEX IF EXISTS idx_entries_account_created;
-- +goose StatementEnd
