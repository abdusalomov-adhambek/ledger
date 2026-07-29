package idempotency

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Repository interface {
	Create(ctx context.Context, idempotency *Idempotency, tx pgx.Tx) error
	GetByTransactionId(ctx context.Context, idempotency *Idempotency) error
}
