package account

import (
	"context"
	"ledger/internal/adapter/errors"
	"ledger/internal/domain/account"

	"log/slog"
)

type AccountApplication struct {
	repo   account.Repository
	logger *slog.Logger
}

func NewAccountApplication(repo account.Repository, logger *slog.Logger) *AccountApplication {
	return &AccountApplication{
		repo:   repo,
		logger: logger,
	}
}

func (a *AccountApplication) CreateAccount(ctx context.Context, req CreateAccountRequest) (*CreateAccountResponse, error) {
	a.logger.Info("AccountApplication.CreateAccount")

	switch req.Currency {
	case account.CurrencyUSD, account.CurrencyEUR, account.CurrencyUZS:
	default:
		a.logger.Error("invalid currency")
		return nil, errors.ErrInvalidCurrency
	}

	newAccount := account.NewAccount("", req.OwnerID, req.Currency, req.Balance, account.AccountStatusOpen)
	if err := a.repo.Create(ctx, newAccount); err != nil {
		a.logger.Error("failed to create account", "error", err)
		return nil, err
	}

	a.logger.Info("account created", "id", newAccount.ID())
	return &CreateAccountResponse{ID: newAccount.ID()}, nil
}

func (a *AccountApplication) GetByID(ctx context.Context, req GetByIDRequest) (*GetByIDResponse, error) {
	a.logger.Info("AccountApplication.GetByID", "id", req.ID)

	account, err := a.repo.GetByID(ctx, req.ID)
	if err != nil {
		a.logger.Error("failed to get account by id", "error", err)
		return nil, err
	}

	res := &GetByIDResponse{
		ID:        account.ID(),
		OwnerID:   account.OwnerID(),
		Currency:  account.Currency(),
		Balance:   account.Balance(),
		CreatedAt: account.CreatedAt().Format("2006-01-02 / 15:04:05"),
		Status:    account.Status(),
	}
	return res, nil
}
