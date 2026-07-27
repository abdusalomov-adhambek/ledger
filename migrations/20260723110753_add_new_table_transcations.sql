-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE TYPE transaction_status AS ENUM ('PENDING', 'COMPLETED', 'CANCELLED', 'REVERSED');

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reference TEXT,
    status transaction_status NOT NULL,
    description TEXT,
    metadata JSONB DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TYPE IF EXISTS transaction_status;
DROP TABLE IF EXISTS transactions;
-- +goose StatementEnd
--
