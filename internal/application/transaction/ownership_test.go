package transaction

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	apperrors "ledger/internal/adapter/errors"
	"ledger/internal/auth"
	"ledger/internal/domain/account"
	"ledger/internal/domain/entry"
	"ledger/internal/domain/transaction"

	"github.com/jackc/pgx/v5"
)

type fakeAccountRepo struct {
	accounts map[string]*account.Account
}

func (f *fakeAccountRepo) Create(context.Context, *account.Account) error { return nil }

func (f *fakeAccountRepo) GetByID(_ context.Context, id string) (*account.Account, error) {
	if a, ok := f.accounts[id]; ok {
		return a, nil
	}
	return nil, pgx.ErrNoRows
}

func (f *fakeAccountRepo) GetForUpdate(context.Context, pgx.Tx, string) (*account.Account, error) {
	panic("reverse must be rejected before locking accounts")
}

func (f *fakeAccountRepo) UpdateBalance(context.Context, *account.Account, pgx.Tx) error {
	panic("reverse must be rejected before updating balances")
}

type fakeEntryRepo struct {
	entries []*entry.Entry
}

func (f *fakeEntryRepo) Create(context.Context, pgx.Tx, *entry.Entry) error { return nil }

func (f *fakeEntryRepo) GetList(context.Context, *entry.Filter, string) ([]*entry.Entry, int64, error) {
	return nil, 0, nil
}

func (f *fakeEntryRepo) GetEntriesByTransactionID(context.Context, string) ([]*entry.Entry, error) {
	return f.entries, nil
}

func TestReverseTransactionChecksDebitedAccountOwner(t *testing.T) {
	accounts := map[string]*account.Account{
		"acc-a": account.NewAccount("acc-a", "owner-a", account.CurrencyUSD, 50, account.AccountStatusOpen),
		"acc-b": account.NewAccount("acc-b", "owner-b", account.CurrencyUSD, 150, account.AccountStatusOpen),
	}
	transfer := []*entry.Entry{
		entry.NewEntry("e-1", "tx-1", "acc-a", entry.EntryTypeDebit, 50),
		entry.NewEntry("e-2", "tx-1", "acc-b", entry.EntryTypeCredit, 50),
	}
	creditOnly := []*entry.Entry{entry.NewEntry("e-2", "tx-1", "acc-b", entry.EntryTypeCredit, 50)}

	tests := []struct {
		name    string
		ctx     context.Context
		entries []*entry.Entry
		wantErr error
	}{
		{"receiver of the transfer", auth.WithSubject(context.Background(), "owner-b"), transfer, apperrors.ErrForbidden},
		{"stranger", auth.WithSubject(context.Background(), "owner-c"), transfer, apperrors.ErrForbidden},
		{"no caller", context.Background(), transfer, apperrors.ErrUnauthenticated},
		{"no debit entry", auth.WithSubject(context.Background(), "owner-b"), creditOnly, transaction.ErrNoEntriesFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// txManager stays nil: every case here must be decided before a transaction opens.
			app := &TransactionApplication{
				logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
				entryRepo:   &fakeEntryRepo{entries: tt.entries},
				accountRepo: &fakeAccountRepo{accounts: accounts},
			}

			err := app.ReverseTransaction(tt.ctx, &ReverseTransactionRequest{TransactionID: "tx-1"})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
