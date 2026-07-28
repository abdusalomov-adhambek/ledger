package account

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Repository interface {
	Create(ctx context.Context, account *Account) error
	GetByID(ctx context.Context, id string) (*Account, error)
	GetForUpdate(ctx context.Context, tx pgx.Tx, id string) (*Account, error)
	UpdateBalance(ctx context.Context, account *Account, tx pgx.Tx) error
}
