package account

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	apperrors "ledger/internal/adapter/errors"
	"ledger/internal/domain/account"

	"github.com/jackc/pgx/v5"
)

// fakeRepo captures the account passed to Create.
type fakeRepo struct {
	created *account.Account
}

func (f *fakeRepo) Create(_ context.Context, a *account.Account) error {
	a.SetID("acc-1")
	f.created = a
	return nil
}

func (f *fakeRepo) GetByID(context.Context, string) (*account.Account, error) {
	return nil, nil
}

func (f *fakeRepo) GetForUpdate(context.Context, pgx.Tx, string) (*account.Account, error) {
	return nil, nil
}

func (f *fakeRepo) UpdateBalance(context.Context, *account.Account, pgx.Tx) error {
	return nil
}

func newTestApp(repo account.Repository) *AccountApplication {
	return NewAccountApplication(repo, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestCreateAccountIgnoresClientBalance(t *testing.T) {
	tests := []struct {
		name    string
		balance int64
	}{
		{"huge balance", 1_000_000_000},
		{"negative balance", -500},
		{"no balance", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{}
			app := newTestApp(repo)

			res, err := app.CreateAccount(context.Background(), CreateAccountRequest{
				OwnerID:  "owner-1",
				Currency: account.CurrencyUSD,
				Balance:  tt.balance,
			})
			if err != nil {
				t.Fatalf("CreateAccount: %v", err)
			}
			if res.ID != "acc-1" {
				t.Fatalf("id = %q, want acc-1", res.ID)
			}
			if got := repo.created.Balance(); got != 0 {
				t.Fatalf("balance = %d, want 0", got)
			}
		})
	}
}

func TestCreateAccountInvalidCurrency(t *testing.T) {
	repo := &fakeRepo{}
	app := newTestApp(repo)

	_, err := app.CreateAccount(context.Background(), CreateAccountRequest{
		OwnerID:  "owner-1",
		Currency: "XXX",
	})
	if !errors.Is(err, apperrors.ErrInvalidCurrency) {
		t.Fatalf("err = %v, want ErrInvalidCurrency", err)
	}
	if repo.created != nil {
		t.Fatal("account was created with an invalid currency")
	}
}
