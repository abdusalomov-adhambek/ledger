-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TYPE entry_type AS ENUM ('DEBIT', 'CREDIT');

CREATE TABLE IF NOT EXISTS entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID REFERENCES transactions(id),
    account_id UUID REFERENCES accounts(id),
    type entry_type NOT NULL,
    amount BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TYPE IF EXISTS entry_type;
DROP TABLE IF EXISTS entries;
-- +goose StatementEnd
