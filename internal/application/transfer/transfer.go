package transfer

import (
	"context"
	"ledger/internal/db"
	"ledger/internal/domain/account"
	"ledger/internal/domain/entry"
	"ledger/internal/domain/transaction"
	"ledger/internal/domain/transfer"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TransferApp struct {
	logger    *slog.Logger
	txManager *db.TxManager

	accountRepo     account.Repository
	transactionRepo transaction.Repository
	entryRepo       entry.Repository
}

func NewTransferApplication(logger *slog.Logger, txManager *db.TxManager, accountRepo account.Repository, transactionRepo transaction.Repository, entryRepo entry.Repository) *TransferApp {
	return &TransferApp{
		logger:          logger,
		txManager:       txManager,
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		entryRepo:       entryRepo,
	}
}

func (t *TransferApp) Transfer(ctx context.Context, req TransferRequest) error {
	if req.Amount <= 0 {
		return transfer.ErrInvalidAmount
	}

	if req.FromAccountID == req.ToAccountID {
		return transfer.ErrSameAccount
	}

	return t.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		// 1. Get `from` account
		from, err := t.accountRepo.GetForUpdate(ctx, tx, req.FromAccountID)
		if err != nil {
			t.logger.Error("failed to get from account", "error", err)
			return err
		}

		// 2. Get `to` account
		to, err := t.accountRepo.GetForUpdate(ctx, tx, req.ToAccountID)
		if err != nil {
			t.logger.Error("failed to get to account", "error", err)
			return err
		}

		// 3. Currency bir xilmi
		if to.Currency() != from.Currency() {
			return transfer.ErrCurrencyMismatch
		}

		// 4. From balance -= amount
		if err = from.Withdraw(req.Amount); err != nil {
			t.logger.Error("failed to withdraw from account", "error", err)
			return err
		}

		// 5. To balance += amount
		if err = to.Deposit(req.Amount); err != nil {
			t.logger.Error("failed to deposit to account", "error", err)
			return err
		}

		// 6. Transaction yaratish
		reference := uuid.NewString()
		tr := transaction.NewTransaction(
			"",
			reference,
			transaction.StatusCompleted,
			req.Description,
			nil,
		)

		if err := t.transactionRepo.Create(ctx, tx, tr); err != nil {
			t.logger.Error("failed to create transaction", "error", err)
			return err
		}

		// 7. Debit entry
		debit := entry.NewEntry(
			"",
			tr.ID(),
			from.ID(),
			entry.EntryTypeDebit,
			req.Amount,
		)
		if err := t.entryRepo.Create(ctx, tx, debit); err != nil {
			t.logger.Error("failed to create debit entry", "error", err)
			return err
		}

		// 8. Credit entry
		credit := entry.NewEntry(
			"",
			tr.ID(),
			to.ID(),
			entry.EntryTypeCredit,
			req.Amount,
		)
		if err := t.entryRepo.Create(ctx, tx, credit); err != nil {
			t.logger.Error("failed to create credit entry", "error", err)
			return err
		}

		// Update account balances
		if err := t.accountRepo.Update(ctx, from, tx); err != nil {
			t.logger.Error("failed to update from account balance", "error", err)
			return err
		}

		if err := t.accountRepo.Update(ctx, to, tx); err != nil {
			t.logger.Error("failed to update to account balance", "error", err)
			return err
		}

		return nil
	})
}
