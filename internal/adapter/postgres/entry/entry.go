package entry

import (
	"context"
	"ledger/internal/domain/account"
	"ledger/internal/domain/entry"
	"ledger/internal/domain/transaction"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EntryRepo struct {
	pgx *pgxpool.Pool
}

func NewEntryRepo(pgx *pgxpool.Pool) *EntryRepo {
	return &EntryRepo{pgx: pgx}
}

func (r *EntryRepo) Create(ctx context.Context, tx pgx.Tx, entry *entry.Entry) error {
	query := `
		INSERT INTO entries (
			transaction_id,
			account_id,
			type,
			amount
		)
		VALUES ($1, $2, $3, $4)
	`
	_, err := tx.Exec(ctx, query, entry.TransactionID(), entry.AccountID(), entry.EntryType(), entry.Amount())
	return err
}

func (r *EntryRepo) GetList(ctx context.Context, filter *entry.Filter) ([]*entry.Entry, int64, error) {
	query := `
		SELECT e.id, e.transaction_id, e.type, e.amount, t.description
		FROM entries e
		INNER JOIN transactions t ON e.transaction_id = t.id
		WHERE e.account_id = $1
	`
	log.Printf("get list", "account_id", filter.AccountID)
	rows, err := r.pgx.Query(ctx, query, &filter.AccountID)
	if err != nil {
		return nil, 0, err
	}

	var data []*entry.Entry

	for rows.Next() {
		var (
			id,
			transactionId,
			entryType,
			description string
			amount int64
		)

		if err := rows.Scan(&id, &transactionId, &entryType, &amount, &description); err != nil {
			return nil, 0, err
		}

		newEntry := entry.NewEntry(id, transactionId, "", entryType, amount)
		newTransaction := transaction.NewTransaction("", "", "", description, nil)
		newEntry.SetTransaction(newTransaction)
		data = append(data, newEntry)
	}

	var count int64
	query = `
		SELECT COUNT(*) FROM entries e
		INNER JOIN transactions t ON e.transaction_id = t.id
		WHERE e.account_id = $1
	`
	if err := r.pgx.QueryRow(ctx, query, &filter.AccountID).Scan(&count); err != nil {
		return nil, 0, err
	}
	return data, count, nil
}

func (r *EntryRepo) GetEntriesByTransactionID(ctx context.Context, transactionId string) ([]*entry.Entry, error) {
	query := `
		SELECT e.id, e.transaction_id, e.type, e.amount, e.account_id, a.balance, a.status
		FROM entries e
		INNER JOIN transactions t ON e.transaction_id = t.id
		INNER JOIN accounts a ON e.account_id = a.id
		WHERE e.transaction_id = $1
	`
	rows, err := r.pgx.Query(ctx, query, &transactionId)
	if err != nil {
		return nil, err
	}

	var data []*entry.Entry
	for rows.Next() {
		var (
			id,
			entryType string
			amount    int64
			accountId string
			balance   int64
			status    string
		)

		if err := rows.Scan(&id, &transactionId, &entryType, &amount, &accountId, &balance, &status); err != nil {
			return nil, err
		}

		newEntry := entry.NewEntry(id, transactionId, accountId, entryType, amount)
		newAccount := account.NewAccount(accountId, "", "", balance, status)
		newEntry.SetAccount(newAccount)
		data = append(data, newEntry)
	}

	return data, nil
}
