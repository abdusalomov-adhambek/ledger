package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Manager interface {
	WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type TxManager struct {
	db *pgxpool.Pool
}

func NewTxManager(db *pgxpool.Pool) *TxManager {
	return &TxManager{
		db: db,
	}
}

func (m *TxManager) WithTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
