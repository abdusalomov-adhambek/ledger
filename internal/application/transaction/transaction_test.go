package transaction

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"sync"
	"testing"

	accountrepo "ledger/internal/adapter/postgres/account"
	entryrepo "ledger/internal/adapter/postgres/entry"
	transactionrepo "ledger/internal/adapter/postgres/transaction"
	"ledger/internal/auth"
	"ledger/internal/db"
	"ledger/internal/domain/transaction"

	"github.com/jackc/pgx/v5/pgxpool"
)

// These tests need a migrated Postgres; set LEDGER_TEST_DATABASE_URL to run them.
type fixture struct {
	pool          *pgxpool.Pool
	app           *TransactionApplication
	fromID, toID  string
	ownerID       string
	transactionID string
}

// newFixture seeds a completed transfer of amount from `from` to `to`, with the given balances after it.
func newFixture(t *testing.T, fromBalance, toBalance, amount int64) *fixture {
	t.Helper()

	url := os.Getenv("LEDGER_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("LEDGER_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	f := &fixture{pool: pool}
	for _, p := range []struct {
		id      *string
		balance int64
	}{{&f.fromID, fromBalance}, {&f.toID, toBalance}} {
		err := pool.QueryRow(ctx, `
			INSERT INTO accounts (owner_id, currency, balance, status)
			VALUES (gen_random_uuid(), 'USD', $1, 'OPEN') RETURNING id`, p.balance).Scan(p.id)
		if err != nil {
			t.Fatalf("seed account: %v", err)
		}
	}

	err = pool.QueryRow(ctx, `INSERT INTO transactions (status) VALUES ('COMPLETED') RETURNING id`).Scan(&f.transactionID)
	if err != nil {
		t.Fatalf("seed transaction: %v", err)
	}

	for _, e := range []struct{ account, kind string }{{f.fromID, "DEBIT"}, {f.toID, "CREDIT"}} {
		if _, err := pool.Exec(ctx, `
			INSERT INTO entries (transaction_id, account_id, type, amount) VALUES ($1, $2, $3, $4)`,
			f.transactionID, e.account, e.kind, amount); err != nil {
			t.Fatalf("seed entry: %v", err)
		}
	}

	if err := pool.QueryRow(ctx, `SELECT owner_id FROM accounts WHERE id = $1`, f.fromID).Scan(&f.ownerID); err != nil {
		t.Fatalf("read owner: %v", err)
	}

	t.Cleanup(func() {
		ctx := context.Background()
		pool.Exec(ctx, `DELETE FROM entries WHERE account_id IN ($1, $2)`, f.fromID, f.toID)
		pool.Exec(ctx, `DELETE FROM transactions WHERE reversal_of = $1`, f.transactionID)
		pool.Exec(ctx, `DELETE FROM transactions WHERE id = $1`, f.transactionID)
		pool.Exec(ctx, `DELETE FROM accounts WHERE id IN ($1, $2)`, f.fromID, f.toID)
	})

	f.app = NewTransactionApplication(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		transactionrepo.NewTransactionRepo(pool),
		entryrepo.NewEntryRepo(pool),
		db.NewTxManager(pool),
		accountrepo.NewAccountRepo(pool),
	)

	return f
}

// ctx is a context for the owner of the debited account, who is allowed to reverse.
func (f *fixture) ctx() context.Context {
	return auth.WithSubject(context.Background(), f.ownerID)
}

func (f *fixture) balances(t *testing.T) (from, to int64) {
	t.Helper()

	ctx := context.Background()
	if err := f.pool.QueryRow(ctx, `SELECT balance FROM accounts WHERE id = $1`, f.fromID).Scan(&from); err != nil {
		t.Fatalf("read from balance: %v", err)
	}
	if err := f.pool.QueryRow(ctx, `SELECT balance FROM accounts WHERE id = $1`, f.toID).Scan(&to); err != nil {
		t.Fatalf("read to balance: %v", err)
	}

	return from, to
}

func (f *fixture) status(t *testing.T) string {
	t.Helper()

	var status string
	if err := f.pool.QueryRow(context.Background(), `SELECT status FROM transactions WHERE id = $1`, f.transactionID).Scan(&status); err != nil {
		t.Fatalf("read status: %v", err)
	}

	return status
}

// assertOneReversal checks that exactly one reversal exists, with its two entries.
func (f *fixture) assertOneReversal(t *testing.T) {
	t.Helper()

	ctx := context.Background()
	var reversals, entries int
	if err := f.pool.QueryRow(ctx, `SELECT count(*) FROM transactions WHERE reversal_of = $1`, f.transactionID).Scan(&reversals); err != nil {
		t.Fatalf("count reversals: %v", err)
	}
	if reversals != 1 {
		t.Fatalf("reversals = %d, want 1", reversals)
	}
	if err := f.pool.QueryRow(ctx, `SELECT count(*) FROM entries WHERE account_id IN ($1, $2)`, f.fromID, f.toID).Scan(&entries); err != nil {
		t.Fatalf("count entries: %v", err)
	}
	if entries != 4 {
		t.Fatalf("entries = %d, want 4 (2 original + 2 reversal)", entries)
	}
}

func TestReverseTransactionTwiceIsRejected(t *testing.T) {
	f := newFixture(t, 50, 150, 50)
	ctx := f.ctx()
	req := &ReverseTransactionRequest{TransactionID: f.transactionID}

	if err := f.app.ReverseTransaction(ctx, req); err != nil {
		t.Fatalf("first reverse: %v", err)
	}
	if from, to := f.balances(t); from != 100 || to != 100 {
		t.Fatalf("after first reverse got balances %d/%d, want 100/100", from, to)
	}

	if err := f.app.ReverseTransaction(ctx, req); !errors.Is(err, transaction.ErrAlreadyReversed) {
		t.Fatalf("second reverse error = %v, want ErrAlreadyReversed", err)
	}
	if from, to := f.balances(t); from != 100 || to != 100 {
		t.Fatalf("after second reverse got balances %d/%d, want 100/100", from, to)
	}
	if got := f.status(t); got != transaction.StatusReversed {
		t.Fatalf("status = %s, want %s", got, transaction.StatusReversed)
	}
	f.assertOneReversal(t)
}

func TestReverseTransactionConcurrentOnlyOneSucceeds(t *testing.T) {
	f := newFixture(t, 50, 150, 50)
	req := &ReverseTransactionRequest{TransactionID: f.transactionID}

	const workers = 8
	errs := make([]error, workers)
	var wg sync.WaitGroup
	for i := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = f.app.ReverseTransaction(f.ctx(), req)
		}()
	}
	wg.Wait()

	succeeded := 0
	for _, err := range errs {
		switch {
		case err == nil:
			succeeded++
		case !errors.Is(err, transaction.ErrAlreadyReversed):
			t.Errorf("unexpected error: %v", err)
		}
	}
	if succeeded != 1 {
		t.Fatalf("%d reversals succeeded, want exactly 1", succeeded)
	}
	if from, to := f.balances(t); from != 100 || to != 100 {
		t.Fatalf("got balances %d/%d, want 100/100", from, to)
	}
	if got := f.status(t); got != transaction.StatusReversed {
		t.Fatalf("status = %s, want %s", got, transaction.StatusReversed)
	}
	f.assertOneReversal(t)
}

func TestReverseTransactionWithoutEntries(t *testing.T) {
	f := newFixture(t, 50, 150, 50)

	var missingID string
	if err := f.pool.QueryRow(context.Background(), `SELECT gen_random_uuid()`).Scan(&missingID); err != nil {
		t.Fatalf("generate id: %v", err)
	}

	err := f.app.ReverseTransaction(f.ctx(), &ReverseTransactionRequest{TransactionID: missingID})
	if !errors.Is(err, transaction.ErrNoEntriesFound) {
		t.Fatalf("error = %v, want ErrNoEntriesFound", err)
	}
}

func TestReverseTransactionInsufficientFundsRollsBack(t *testing.T) {
	// the destination spent part of the amount, so it cannot give it back
	f := newFixture(t, 50, 10, 50)

	err := f.app.ReverseTransaction(f.ctx(), &ReverseTransactionRequest{TransactionID: f.transactionID})
	if err == nil {
		t.Fatal("reverse succeeded, want insufficient balance error")
	}
	if from, to := f.balances(t); from != 50 || to != 10 {
		t.Fatalf("got balances %d/%d, want 50/10", from, to)
	}
	if got := f.status(t); got != transaction.StatusCompleted {
		t.Fatalf("status = %s, want %s", got, transaction.StatusCompleted)
	}
}

func TestReverseTransactionReversalIsNotReversible(t *testing.T) {
	f := newFixture(t, 50, 150, 50)
	ctx := f.ctx()

	if err := f.app.ReverseTransaction(ctx, &ReverseTransactionRequest{TransactionID: f.transactionID}); err != nil {
		t.Fatalf("reverse: %v", err)
	}

	var reversalID string
	if err := f.pool.QueryRow(ctx, `SELECT id FROM transactions WHERE reversal_of = $1`, f.transactionID).Scan(&reversalID); err != nil {
		t.Fatalf("find reversal: %v", err)
	}

	// the reversal debits the destination account, so its owner is the one allowed to reverse it
	var toOwnerID string
	if err := f.pool.QueryRow(ctx, `SELECT owner_id FROM accounts WHERE id = $1`, f.toID).Scan(&toOwnerID); err != nil {
		t.Fatalf("read owner: %v", err)
	}

	err := f.app.ReverseTransaction(auth.WithSubject(ctx, toOwnerID), &ReverseTransactionRequest{TransactionID: reversalID})
	if !errors.Is(err, transaction.ErrAlreadyReversed) {
		t.Fatalf("reversing a reversal error = %v, want ErrAlreadyReversed", err)
	}
	if from, to := f.balances(t); from != 100 || to != 100 {
		t.Fatalf("got balances %d/%d, want 100/100", from, to)
	}
	f.assertOneReversal(t)
}
