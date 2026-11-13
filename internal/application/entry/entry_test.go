package entry

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

	"github.com/jackc/pgx/v5"
)

type fakeAccountRepo struct {
	stored *account.Account
}

func (f *fakeAccountRepo) Create(context.Context, *account.Account) error { return nil }

func (f *fakeAccountRepo) GetByID(context.Context, string) (*account.Account, error) {
	if f.stored == nil {
		return nil, pgx.ErrNoRows
	}
	return f.stored, nil
}

func (f *fakeAccountRepo) GetForUpdate(context.Context, pgx.Tx, string) (*account.Account, error) {
	return nil, nil
}

func (f *fakeAccountRepo) UpdateBalance(context.Context, *account.Account, pgx.Tx) error { return nil }

type fakeEntryRepo struct {
	listed    bool
	gotFilter *entry.Filter
}

func (f *fakeEntryRepo) Create(context.Context, pgx.Tx, *entry.Entry) error { return nil }

func (f *fakeEntryRepo) GetList(_ context.Context, filter *entry.Filter, _ string) ([]*entry.Entry, int64, error) {
	f.listed = true
	f.gotFilter = filter
	return nil, 0, nil
}

func (f *fakeEntryRepo) GetEntriesByTransactionID(context.Context, string) ([]*entry.Entry, error) {
	return nil, nil
}

func TestGetHistoryListChecksOwner(t *testing.T) {
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
			entries := &fakeEntryRepo{}
			app := NewEntryApplication(slog.New(slog.NewTextHandler(io.Discard, nil)), entries, &fakeAccountRepo{stored: tt.stored})

			limit := int64(5)
			filter := &entry.Filter{Limit: &limit}
			_, _, err := app.GetHistoryList(tt.ctx, filter, &GetHistoryListRequest{AccountID: "acc-1"})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if listed := entries.listed; listed != (tt.wantErr == nil) {
				t.Fatalf("entries queried = %v, want %v", listed, tt.wantErr == nil)
			}
			if tt.wantErr == nil && entries.gotFilter != filter {
				t.Fatal("filter was not passed through unchanged")
			}
		})
	}
}
