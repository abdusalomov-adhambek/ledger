package transfer

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
	"ledger/internal/domain/idempotency"
	"ledger/internal/domain/outbox_events"
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
		panic("transfer must be rejected before locking accounts")
	}
	return f.accounts[id], nil
}

func (f *fakeAccountRepo) UpdateBalance(_ context.Context, a *account.Account, _ pgx.Tx) error {
	if !f.locked {
		panic("transfer must be rejected before updating balances")
	}
	f.updated = append(f.updated, a.ID())
	return nil
}

// fakeIdempotencyRepo reports every key as already used by transaction tx-1.
type fakeIdempotencyRepo struct{}

func (fakeIdempotencyRepo) Create(context.Context, *idempotency.Idempotency, pgx.Tx) error {
	return nil
}

func (fakeIdempotencyRepo) GetByTransactionId(_ context.Context, i *idempotency.Idempotency) error {
	*i = *idempotency.NewIdempotency(i.Key(), "tx-1", nil)
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

func newTestApp(accounts map[string]*account.Account, entries []*entry.Entry, idem idempotency.Repository) *TransferApp {
	// txManager stays nil: every case here must be decided before a transaction opens.
	return &TransferApp{
		logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
		accountRepo:     &fakeAccountRepo{accounts: accounts},
		entryRepo:       &fakeEntryRepo{entries: entries},
		idempotencyRepo: idem,
	}
}

func accounts() map[string]*account.Account {
	return map[string]*account.Account{
		"acc-a": account.NewAccount("acc-a", "owner-a", account.CurrencyUSD, 500, account.AccountStatusOpen),
		"acc-b": account.NewAccount("acc-b", "owner-b", account.CurrencyUSD, 500, account.AccountStatusOpen),
	}
}

func TestTransferChecksFromAccountOwner(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		from    string
		wantErr error
	}{
		{"from account of another user", auth.WithSubject(context.Background(), "owner-b"), "acc-a", apperrors.ErrForbidden},
		{"no caller", context.Background(), "acc-a", apperrors.ErrUnauthenticated},
		{"missing from account looks like another user's", auth.WithSubject(context.Background(), "owner-b"), "acc-x", apperrors.ErrForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApp(accounts(), nil, nil)

			id, err := app.Transfer(tt.ctx, &TransferRequest{FromAccountID: tt.from, ToAccountID: "acc-b", Amount: 100, IdempotencyKey: "key-1"})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if id != "" {
				t.Fatalf("id = %q, want none", id)
			}
		})
	}
}

func TestTransferReplayDoesNotLeakOthersTransaction(t *testing.T) {
	// tx-1 was a transfer out of acc-a; owner-b replays its key with their own account.
	original := []*entry.Entry{
		entry.NewEntry("e-1", "tx-1", "acc-a", entry.EntryTypeDebit, 100),
		entry.NewEntry("e-2", "tx-1", "acc-b", entry.EntryTypeCredit, 100),
	}
	ctx := auth.WithSubject(context.Background(), "owner-b")
	app := newTestApp(accounts(), original, fakeIdempotencyRepo{})

	id, err := app.Transfer(ctx, &TransferRequest{FromAccountID: "acc-b", ToAccountID: "acc-a", Amount: 100, IdempotencyKey: "key-1"})
	if !errors.Is(err, apperrors.ErrForbidden) {
		t.Fatalf("err = %v, want ErrForbidden", err)
	}
	if id != "" {
		t.Fatalf("id = %q, want none", id)
	}
}

func TestTransferReplayByOwnerReturnsTransaction(t *testing.T) {
	original := []*entry.Entry{
		entry.NewEntry("e-1", "tx-1", "acc-a", entry.EntryTypeDebit, 100),
		entry.NewEntry("e-2", "tx-1", "acc-b", entry.EntryTypeCredit, 100),
	}
	ctx := auth.WithSubject(context.Background(), "owner-a")
	app := newTestApp(accounts(), original, fakeIdempotencyRepo{})

	id, err := app.Transfer(ctx, &TransferRequest{FromAccountID: "acc-a", ToAccountID: "acc-b", Amount: 100, IdempotencyKey: "key-1"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if id != "tx-1" {
		t.Fatalf("id = %q, want tx-1", id)
	}
}

// freshIdempotencyRepo reports every key as unused.
type freshIdempotencyRepo struct{}

func (freshIdempotencyRepo) Create(context.Context, *idempotency.Idempotency, pgx.Tx) error {
	return nil
}

func (freshIdempotencyRepo) GetByTransactionId(context.Context, *idempotency.Idempotency) error {
	return pgx.ErrNoRows
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

type fakeOutboxRepo struct{}

func (fakeOutboxRepo) Create(context.Context, pgx.Tx, *outbox_events.OutboxEvent) error { return nil }

func (fakeOutboxRepo) GetPending(context.Context) ([]*outbox_events.OutboxEvent, error) {
	return nil, nil
}

func (fakeOutboxRepo) MarkProcessed(context.Context, string, string) error { return nil }

// fakeTxManager runs the function without a database transaction.
type fakeTxManager struct{}

func (fakeTxManager) WithTransaction(_ context.Context, fn func(tx pgx.Tx) error) error {
	return fn(nil)
}

func TestTransferByOwnerMovesMoney(t *testing.T) {
	accs := accounts()
	accountRepo := &fakeAccountRepo{accounts: accs, locked: true}
	app := &TransferApp{
		logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
		txManager:       fakeTxManager{},
		accountRepo:     accountRepo,
		transactionRepo: fakeTransactionRepo{},
		entryRepo:       &fakeEntryRepo{},
		idempotencyRepo: freshIdempotencyRepo{},
		outboxEventRepo: fakeOutboxRepo{},
	}

	// owner-a pays into acc-b, which belongs to someone else
	ctx := auth.WithSubject(context.Background(), "owner-a")
	id, err := app.Transfer(ctx, &TransferRequest{FromAccountID: "acc-a", ToAccountID: "acc-b", Amount: 100, IdempotencyKey: "key-1"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if id != "tx-new" {
		t.Fatalf("id = %q, want tx-new", id)
	}
	if from, to := accs["acc-a"].Balance(), accs["acc-b"].Balance(); from != 400 || to != 600 {
		t.Fatalf("balances = %d/%d, want 400/600", from, to)
	}
	if len(accountRepo.updated) != 2 {
		t.Fatalf("updated accounts = %v, want both", accountRepo.updated)
	}
}
