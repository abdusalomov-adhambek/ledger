package entry

import (
	"context"
	"os"
	"testing"

	"ledger/internal/domain/entry"

	"github.com/jackc/pgx/v5/pgxpool"
)

// newTestRepo connects to the database in TEST_POSTGRES_URL (migrations applied)
// and skips the test when it is not set.
func newTestRepo(t *testing.T) (*EntryRepo, *pgxpool.Pool) {
	t.Helper()
	url := os.Getenv("TEST_POSTGRES_URL")
	if url == "" {
		t.Skip("TEST_POSTGRES_URL is not set")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return NewEntryRepo(pool), pool
}

// seedEntry creates an account with one entry and returns the account id.
func seedEntry(t *testing.T, pool *pgxpool.Pool, description string) string {
	t.Helper()
	ctx := context.Background()
	var accountID, transactionID string
	err := pool.QueryRow(ctx,
		`INSERT INTO accounts (owner_id, currency) VALUES (gen_random_uuid(), 'USD') RETURNING id`,
	).Scan(&accountID)
	if err != nil {
		t.Fatalf("seed account: %v", err)
	}
	err = pool.QueryRow(ctx,
		`INSERT INTO transactions (status, description) VALUES ('COMPLETED', $1) RETURNING id`, description,
	).Scan(&transactionID)
	if err != nil {
		t.Fatalf("seed transaction: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO entries (transaction_id, account_id, type, amount) VALUES ($1, $2, 'CREDIT', 100)`,
		transactionID, accountID,
	); err != nil {
		t.Fatalf("seed entry: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(ctx, `DELETE FROM entries WHERE account_id = $1`, accountID)
		pool.Exec(ctx, `DELETE FROM transactions WHERE id = $1`, transactionID)
		pool.Exec(ctx, `DELETE FROM accounts WHERE id = $1`, accountID)
	})
	return accountID
}

func TestGetListValidAccount(t *testing.T) {
	repo, pool := newTestRepo(t)
	accountID := seedEntry(t, pool, "mine")
	seedEntry(t, pool, "other")

	got, count, err := repo.GetList(context.Background(), &entry.Filter{}, accountID)
	if err != nil {
		t.Fatalf("GetList: %v", err)
	}
	if len(got) != 1 || count != 1 {
		t.Fatalf("got %d entries, count %d; want 1, 1", len(got), count)
	}
	if got[0].Transaction().Description() != "mine" {
		t.Fatalf("got description %q", got[0].Transaction().Description())
	}
}

func TestGetListAccountIDInjection(t *testing.T) {
	repo, pool := newTestRepo(t)
	accountID := seedEntry(t, pool, "mine")
	seedEntry(t, pool, "other")

	// With the id interpolated into the SQL this becomes
	// `account_id = '<id>' OR '1'='1'` and returns every account's entries.
	payload := accountID + "' OR '1'='1"
	got, _, err := repo.GetList(context.Background(), &entry.Filter{}, payload)
	if err == nil {
		t.Fatalf("expected an invalid uuid error, got %d entries", len(got))
	}
	if len(got) != 0 {
		t.Fatalf("got %d entries for an injected account id; want 0", len(got))
	}
}
