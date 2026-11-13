package auth

import (
	"context"
	"errors"

	apperrors "ledger/internal/adapter/errors"
	"ledger/internal/domain/account"

	"github.com/jackc/pgx/v5"
)

type subjectKey struct{}

// WithSubject returns a context that carries the authenticated caller's id.
func WithSubject(ctx context.Context, subject string) context.Context {
	return context.WithValue(ctx, subjectKey{}, subject)
}

// SubjectFromContext returns the authenticated caller's id, if there is one.
func SubjectFromContext(ctx context.Context) (string, bool) {
	subject, ok := ctx.Value(subjectKey{}).(string)
	return subject, ok && subject != ""
}

// LoadOwnedAccount loads the account and checks that the caller owns it.
// It fails with ErrUnauthenticated when the context has no caller, and with
// ErrForbidden both when the account is another user's and when it does not
// exist, so that ids cannot be probed.
func LoadOwnedAccount(ctx context.Context, repo account.Repository, id string) (*account.Account, error) {
	subject, ok := SubjectFromContext(ctx)
	if !ok {
		return nil, apperrors.ErrUnauthenticated
	}

	acc, err := repo.GetByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.ErrForbidden
	}
	if err != nil {
		return nil, err
	}

	if acc.OwnerID() != subject {
		return nil, apperrors.ErrForbidden
	}

	return acc, nil
}
