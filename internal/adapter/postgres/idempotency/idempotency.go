package idempotency

import (
	"context"
	"errors"
	"ledger/internal/domain/idempotency"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IdempotencyRepo struct {
	pgx *pgxpool.Pool
}

func NewIdempotencyRepo(pgx *pgxpool.Pool) *IdempotencyRepo {
	return &IdempotencyRepo{pgx: pgx}
}

func (r *IdempotencyRepo) Create(ctx context.Context, idempotency *idempotency.Idempotency, tx pgx.Tx) error {

	query := `INSERT INTO idempotency_keys (
			key,
			transaction_id,
			created_at
		) VALUES ($1, $2, $3)
		RETURNING key
		`

	var key string
	if err := tx.QueryRow(ctx, query, idempotency.Key(), idempotency.TransactionID(), idempotency.CreatedAt()).Scan(&key); err != nil {
		return err
	}

	idempotency.SetKey(key)
	return nil
}

func (r *IdempotencyRepo) GetByTransactionId(ctx context.Context, idempotency *idempotency.Idempotency) error {
	query := `SELECT transaction_id FROM idempotency_keys WHERE key = $1`

	var trsactionId string

	if err := r.pgx.QueryRow(ctx, query, idempotency.Key()).Scan(&trsactionId); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("idempotency not found")
		}
		return err
	}

	idempotency.SetTransactionID(trsactionId)

	return nil
}
