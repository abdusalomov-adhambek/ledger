package transaction

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Repository interface {
	Create(ctx context.Context, tx pgx.Tx, transaction *Transaction) error
	Reverse(ctx context.Context, transactionId string, tx pgx.Tx) error
	UpdateReversalOf(ctx context.Context, transaction *Transaction, transactionId string, tx pgx.Tx) error
}
