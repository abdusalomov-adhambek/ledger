package transaction

import (
	"context"
	"ledger/internal/db"
	"ledger/internal/domain/account"
	"ledger/internal/domain/entry"
	"ledger/internal/domain/transaction"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

type TransactionApplication struct {
	logger      *slog.Logger
	repo        transaction.Repository
	entryRepo   entry.Repository
	accountRepo account.Repository

	txManager *db.TxManager
}

func NewTransactionApplication(logger *slog.Logger, repo transaction.Repository, entryRepo entry.Repository, txManager *db.TxManager, accountRepo account.Repository) *TransactionApplication {
	return &TransactionApplication{
		logger:      logger,
		repo:        repo,
		entryRepo:   entryRepo,
		txManager:   txManager,
		accountRepo: accountRepo,
	}
}

func (t *TransactionApplication) ReverseTransaction(ctx context.Context, req *ReverseTransactionRequest) error {
	t.logger.Info("transactionApplication.ReverseTransaction", "transaction_id", req.TransactionID)

	entries, err := t.entryRepo.GetEntriesByTransactionID(ctx, req.TransactionID)
	if err != nil {
		t.logger.Error("failed to get entries by transaction id", "transactionApplication.ReverseTransaction", req.TransactionID, "error", err)
		return err
	}

	// check if entries are found
	if len(entries) == 0 {
		return transaction.ErrNoEntriesFound
	}

	return t.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {

		// get entries and accounts
		fromAccount := &account.Account{}
		toAccount := &account.Account{}
		for _, entry := range entries {
			switch entry.EntryType() {
			case "DEBIT":
				fromAccount = entry.Account()
			case "CREDIT":
				toAccount = entry.Account()
			}
		}

		//deposit amount back to fromAccount
		if err := fromAccount.Deposit(entries[0].Amount()); err != nil {
			t.logger.Error("failed to withdraw from account", "transactionApplication.ReverseTransaction", req.TransactionID, "error", err)
			return err
		}

		//withdraw amount from toAccount
		if err := toAccount.Withdraw(entries[0].Amount()); err != nil {
			t.logger.Error("failed to deposit to account", "transactionApplication.ReverseTransaction", req.TransactionID, "error", err)
			return err
		}

		// update fromAccount balance
		if err := t.accountRepo.UpdateBalance(ctx, fromAccount, tx); err != nil {
			t.logger.Error("failed to update from account balance", "transactionApplication.ReverseTransaction", req.TransactionID, "error", err)
			return err
		}

		// update toAccount balance
		if err := t.accountRepo.UpdateBalance(ctx, toAccount, tx); err != nil {
			t.logger.Error("failed to update to account balance", "transactionApplication.ReverseTransaction", req.TransactionID, "error", err)
			return err
		}

		// Create new transaction
		newTransaction := transaction.NewTransaction("", "", transaction.StatusCompleted, "", nil)
		if err := t.repo.Create(ctx, tx, newTransaction); err != nil {
			t.logger.Error("failed to create transaction", "transactionApplication.ReverseTransaction", req.TransactionID, "error", err)
			return err
		}

		// Update reversal of original transaction
		if err := t.repo.UpdateReversalOf(ctx, newTransaction, req.TransactionID, tx); err != nil {
			t.logger.Error("failed to update reversal of", "transactionApplication.ReverseTransaction", req.TransactionID, "error", err)
			return err
		}

		// Update original transaction status to reversed
		if err := t.repo.Reverse(ctx, req.TransactionID, tx); err != nil {
			t.logger.Error("failed to reverse transaction", "transactionApplication.ReverseTransaction", req.TransactionID, "error", err)
			return err
		}

		// Create new entries for reversed transaction
		debit := entry.NewEntry("", newTransaction.ID(), fromAccount.ID(), entry.EntryTypeDebit, entries[0].Amount())
		credit := entry.NewEntry("", newTransaction.ID(), toAccount.ID(), entry.EntryTypeCredit, entries[0].Amount())

		// Create new debit entry
		if err := t.entryRepo.Create(ctx, tx, debit); err != nil {
			t.logger.Error("failed to create debit entry", "transactionApplication.ReverseTransaction", req.TransactionID, "error", err)
			return err
		}

		// Create new credit entry
		if err := t.entryRepo.Create(ctx, tx, credit); err != nil {
			t.logger.Error("failed to create credit entry", "transactionApplication.ReverseTransaction", req.TransactionID, "error", err)
			return err
		}

		return nil
	})
}
