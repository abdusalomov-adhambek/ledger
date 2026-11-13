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
	// locked allows GetForUpdate and UpdateBalance; otherwise they must not be reached.
	locked  bool
	updated []string
}

func (f *fakeAccountRepo) Create(context.Context, *account.Account) error { return nil }

func (f *fakeAccountRepo) GetByID(_ context.Context, id string) (*account.Account, error) {
	if a, ok := f.accounts[id]; ok {
		return a, nil
	}
	return nil, pgx.ErrNoRows
}

func (f *fakeAccountRepo) GetForUpdate(_ context.Context, _ pgx.Tx, id string) (*account.Account, error) {
	if !f.locked {
		panic("reverse must be rejected before locking accounts")
	}
	return f.accounts[id], nil
}

func (f *fakeAccountRepo) UpdateBalance(_ context.Context, a *account.Account, _ pgx.Tx) error {
	if !f.locked {
		panic("reverse must be rejected before updating balances")
	}
	f.updated = append(f.updated, a.ID())
	return nil
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

type fakeTransactionRepo struct{}

func (fakeTransactionRepo) Create(_ context.Context, _ pgx.Tx, tr *transaction.Transaction) error {
	tr.SetID("tx-new")
	return nil
}

func (fakeTransactionRepo) Reverse(context.Context, string, pgx.Tx) error { return nil }

func (fakeTransactionRepo) UpdateReversalOf(context.Context, *transaction.Transaction, string, pgx.Tx) error {
	return nil
}

// fakeTxManager runs the function without a database transaction.
type fakeTxManager struct{}

func (fakeTxManager) WithTransaction(_ context.Context, fn func(tx pgx.Tx) error) error {
	return fn(nil)
}

func TestReverseTransactionByOwnerMovesMoneyBack(t *testing.T) {
	accs := map[string]*account.Account{
		"acc-a": account.NewAccount("acc-a", "owner-a", account.CurrencyUSD, 50, account.AccountStatusOpen),
		"acc-b": account.NewAccount("acc-b", "owner-b", account.CurrencyUSD, 150, account.AccountStatusOpen),
	}
	accountRepo := &fakeAccountRepo{accounts: accs, locked: true}
	app := &TransactionApplication{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		repo:   fakeTransactionRepo{},
		entryRepo: &fakeEntryRepo{entries: []*entry.Entry{
			entry.NewEntry("e-1", "tx-1", "acc-a", entry.EntryTypeDebit, 50),
			entry.NewEntry("e-2", "tx-1", "acc-b", entry.EntryTypeCredit, 50),
		}},
		accountRepo: accountRepo,
		txManager:   fakeTxManager{},
	}

	ctx := auth.WithSubject(context.Background(), "owner-a")
	if err := app.ReverseTransaction(ctx, &ReverseTransactionRequest{TransactionID: "tx-1"}); err != nil {
		t.Fatalf("err = %v", err)
	}
	if from, to := accs["acc-a"].Balance(), accs["acc-b"].Balance(); from != 100 || to != 100 {
		t.Fatalf("balances = %d/%d, want 100/100", from, to)
	}
	if len(accountRepo.updated) != 2 {
		t.Fatalf("updated accounts = %v, want both", accountRepo.updated)
	}
}
