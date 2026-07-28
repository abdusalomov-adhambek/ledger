package transaction

import (
	"context"
	"ledger/internal/domain/transaction"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepo struct {
	pgx *pgxpool.Pool
}

func NewTransactionRepo(pgx *pgxpool.Pool) *TransactionRepo {
	return &TransactionRepo{pgx: pgx}
}

func (t *TransactionRepo) Create(ctx context.Context, tx pgx.Tx, tr *transaction.Transaction) error {
	query := `
	INSERT INTO transactions (
		reference,
		status,
		description
	)
	VALUES ($1, $2, $3)
	RETURNING id
	`

	var id string

	if err := tx.QueryRow(ctx, query,
		tr.Reference(),
		tr.Status(),
		tr.Description()).Scan(&id); err != nil {
		return err
	}

	tr.SetID(id)

	return nil
}

func (t *TransactionRepo) Reverse(ctx context.Context, transactionId string, tx pgx.Tx) error {
	query := `
		UPDATE transactions
		SET status = 'REVERSED'
		WHERE id = $1 AND status = 'COMPLETED'
	`

	if _, err := tx.Exec(ctx, query, transactionId); err != nil {
		return err
	}

	return nil
}

func (t *TransactionRepo) UpdateReversalOf(ctx context.Context, newTransaction *transaction.Transaction, transactionId string, tx pgx.Tx) error {
	query := `
		UPDATE transactions
		SET reversal_of = $1
		WHERE id = $2
	`

	if _, err := tx.Exec(ctx, query, transactionId, newTransaction.ID()); err != nil {
		return err
	}

	return nil
}
