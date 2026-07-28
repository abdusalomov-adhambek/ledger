-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS reversal_of UUID REFERENCES transactions(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE transactions DROP COLUMN IF EXISTS reversal_of;
-- +goose StatementEnd
