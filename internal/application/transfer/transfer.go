package transfer

import (
	"context"
	"encoding/json"
	apperrors "ledger/internal/adapter/errors"
	"ledger/internal/auth"
	"ledger/internal/db"
	"ledger/internal/domain/account"
	"ledger/internal/domain/entry"
	"ledger/internal/domain/idempotency"
	"ledger/internal/domain/outbox_events"
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
	idempotencyRepo idempotency.Repository
	outboxEventRepo outbox_events.Repository
}

func NewTransferApplication(
	logger *slog.Logger,
	txManager *db.TxManager,
	accountRepo account.Repository,
	transactionRepo transaction.Repository,
	entryRepo entry.Repository,
	idempotencyRepo idempotency.Repository,
	outboxEventRepo outbox_events.Repository,
) *TransferApp {
	return &TransferApp{
		logger:          logger,
		txManager:       txManager,
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		entryRepo:       entryRepo,
		idempotencyRepo: idempotencyRepo,
		outboxEventRepo: outboxEventRepo,
	}
}

func (t *TransferApp) Transfer(ctx context.Context, req *TransferRequest) (string, error) {
	var transactionID string

	// only the owner of the from account may move its money
	if _, err := auth.LoadOwnedAccount(ctx, t.accountRepo, req.FromAccountID); err != nil {
		t.logger.Error("failed to get from account", "error", err)
		return "", err
	}

	if req.Amount <= 0 {
		return "", transfer.ErrInvalidAmount
	}

	if req.FromAccountID == req.ToAccountID {
		return "", transfer.ErrSameAccount
	}

	newIdempotency := idempotency.NewIdempotency(req.IdempotencyKey, "", nil)
	if err := t.idempotencyRepo.GetByTransactionId(ctx, newIdempotency); err == nil {
		// the key may belong to a transfer from another account
		if err := t.requireDebitFrom(ctx, newIdempotency.TransactionID(), req.FromAccountID); err != nil {
			return "", err
		}

		return newIdempotency.TransactionID(), nil
	}

	if err := t.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {

		smallID := req.FromAccountID
		largeID := req.ToAccountID

		if smallID > largeID {
			smallID, largeID = largeID, smallID
		}

		// 1. Get `from` account
		firstAccount, err := t.accountRepo.GetForUpdate(ctx, tx, smallID)
		if err != nil {
			t.logger.Error("failed to get from account", "error", err)
			return err
		}

		// 2. Get `to` account
		secondAccount, err := t.accountRepo.GetForUpdate(ctx, tx, largeID)
		if err != nil {
			t.logger.Error("failed to get to account", "error", err)
			return err
		}

		var from, to *account.Account

		if firstAccount.ID() == req.FromAccountID {
			from = firstAccount
			to = secondAccount
		} else {
			from = secondAccount
			to = firstAccount
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

		// 6. Create transaction
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

		transactionID = tr.ID()
		newIdempotency := idempotency.NewIdempotency(req.IdempotencyKey, tr.ID(), nil)
		if err := t.idempotencyRepo.Create(ctx, newIdempotency, tx); err != nil {
			t.logger.Error("failed to create idempotency", "error", err)
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
		if err := t.accountRepo.UpdateBalance(ctx, from, tx); err != nil {
			t.logger.Error("failed to update from account balance", "error", err)
			return err
		}

		if err := t.accountRepo.UpdateBalance(ctx, to, tx); err != nil {
			t.logger.Error("failed to update to account balance", "error", err)
			return err
		}

		// 9. Create outbox event
		payloadMarshal, err := json.Marshal(
			map[string]any{
				"transaction_id":  transactionID,
				"from_account_id": req.FromAccountID,
				"to_account_id":   req.ToAccountID,
				"amount":          req.Amount,
				"currency":        from.Currency(),
				"description":     req.Description,
			},
		)

		if err != nil {
			t.logger.Error("failed to marshal payload", "error", err)
			return err
		}
		outboxEvent := outbox_events.NewOutboxEvent("", outbox_events.EventTransferCompleted, transactionID, payloadMarshal, outbox_events.StatusPending)
		if err := t.outboxEventRepo.Create(ctx, tx, outboxEvent); err != nil {
			t.logger.Error("failed to create outbox event", "error", err)
			return err
		}

		return nil
	}); err != nil {
		return "", err
	}

	return transactionID, nil
}

// requireDebitFrom fails with ErrForbidden unless the transaction debited fromAccountID.
func (t *TransferApp) requireDebitFrom(ctx context.Context, transactionID, fromAccountID string) error {
	entries, err := t.entryRepo.GetEntriesByTransactionID(ctx, transactionID)
	if err != nil {
		t.logger.Error("failed to get entries by transaction id", "transaction_id", transactionID, "error", err)
		return err
	}

	for _, e := range entries {
		if e.EntryType() == entry.EntryTypeDebit && e.AccountID() == fromAccountID {
			return nil
		}
	}

	return apperrors.ErrForbidden
}
