package account

import (
	"context"
	"fmt"
	"ledger/internal/domain/account"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountRepo struct {
	pgx *pgxpool.Pool
}

func NewAccountRepo(pgx *pgxpool.Pool) *AccountRepo {
	return &AccountRepo{pgx: pgx}
}

func (ar *AccountRepo) Create(ctx context.Context, account *account.Account) error {
	query := `
	INSERT INTO accounts
	(owner_id, currency, balance, status)
	VALUES ($1, $2, $3, $4)
	RETURNING id
	`

	var id string
	err := ar.pgx.QueryRow(
		ctx,
		query,
		account.OwnerID(),
		account.Currency(),
		account.Balance(),
		account.Status(),
	).Scan(&id)

	if err != nil {
		return err
	}
	account.SetID(id)

	return nil
}

func (ar *AccountRepo) GetByID(ctx context.Context, id string) (*account.Account, error) {
	query := fmt.Sprintf(`
	SELECT
		owner_id,
		currency,
		balance,
		status,
		created_at
	FROM accounts
	WHERE id = $1`)

	var (
		ownerId   string
		currency  string
		balance   int64
		status    string
		createdAt time.Time
	)

	row := ar.pgx.QueryRow(ctx, query, id)
	err := row.Scan(&ownerId, &currency, &balance, &status, &createdAt)
	if err != nil {
		return nil, err
	}

	newAcc := account.NewAccount(
		id,
		ownerId,
		currency,
		balance,
		status,
	)
	newAcc.SetCreatedAt(createdAt)
	return newAcc, nil
}

func (ar *AccountRepo) GetForUpdate(ctx context.Context, tx pgx.Tx, id string) (*account.Account, error) {
	query := `
		SELECT
		    owner_id,
		    currency,
		    balance,
		    status
		FROM accounts
		WHERE id = $1
		FOR UPDATE
	`

	var (
		ownerId  string
		currency string
		balance  int64
		status   string
	)

	row := tx.QueryRow(ctx, query, id)
	err := row.Scan(&ownerId, &currency, &balance, &status)
	if err != nil {
		return nil, err
	}

	newAcc := account.NewAccount(
		id,
		ownerId,
		currency,
		balance,
		status,
	)
	return newAcc, nil
}

func (ar *AccountRepo) UpdateBalance(ctx context.Context, account *account.Account, tx pgx.Tx) error {
	query := `
		UPDATE accounts
		SET balance = $1
		WHERE id = $2
	`
	_, err := tx.Exec(ctx, query, account.Balance(), account.ID())
	return err
}
