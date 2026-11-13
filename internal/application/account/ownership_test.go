package account

import (
	"context"
	"errors"
	"testing"

	apperrors "ledger/internal/adapter/errors"
	"ledger/internal/auth"
	"ledger/internal/domain/account"

	"github.com/jackc/pgx/v5"
)

// getRepo serves one stored account (or pgx.ErrNoRows).
type getRepo struct {
	fakeRepo
	stored *account.Account
}

func (g *getRepo) GetByID(context.Context, string) (*account.Account, error) {
	if g.stored == nil {
		return nil, pgx.ErrNoRows
	}
	return g.stored, nil
}

func TestGetByIDChecksOwner(t *testing.T) {
	stored := account.NewAccount("acc-1", "owner-1", account.CurrencyUSD, 500, account.AccountStatusOpen)

	tests := []struct {
		name    string
		ctx     context.Context
		stored  *account.Account
		wantErr error
	}{
		{"owner", auth.WithSubject(context.Background(), "owner-1"), stored, nil},
		{"other user", auth.WithSubject(context.Background(), "owner-2"), stored, apperrors.ErrForbidden},
		{"no caller", context.Background(), stored, apperrors.ErrUnauthenticated},
		{"missing account looks like another user's", auth.WithSubject(context.Background(), "owner-2"), nil, apperrors.ErrForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApp(&getRepo{stored: tt.stored})

			res, err := app.GetByID(tt.ctx, GetByIDRequest{ID: "acc-1"})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if res != nil {
					t.Fatalf("response = %+v, want none", res)
				}
				return
			}
			if res.OwnerID != "owner-1" || res.Balance != 500 {
				t.Fatalf("response = %+v", res)
			}
		})
	}
}
