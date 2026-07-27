-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TYPE account_status AS ENUM('OPEN', 'CLOSED', 'BLOCKED');
CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL ,
    currency CHAR(3) NOT NULL,
    balance BIGINT NOT NULL DEFAULT 0,
    status account_status DEFAULT 'OPEN',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TYPE IF EXISTS account_status;
DROP TABLE IF EXISTS accounts;
-- +goose StatementEnd
