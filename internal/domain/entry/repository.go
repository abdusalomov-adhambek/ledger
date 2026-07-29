package entry

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Repository interface {
	Create(ctx context.Context, tx pgx.Tx, entry *Entry) error
	GetList(ctx context.Context, filter *Filter, accountId string) ([]*Entry, int64, error)
	GetEntriesByTransactionID(ctx context.Context, transactionId string) ([]*Entry, error)
}

type Filter struct {
	Limit  *int64
	Offset *int64
	Page   *int64
}
