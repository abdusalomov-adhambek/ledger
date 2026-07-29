package main

import (
	"ledger/internal/adapter/postgres/account"
	"ledger/internal/adapter/postgres/entry"
	"ledger/internal/adapter/postgres/idempotency"
	"ledger/internal/adapter/postgres/transaction"
	"ledger/internal/config"
	"ledger/internal/db"
	"ledger/internal/logger"
	"ledger/internal/server"
	"log"

	accountapplication "ledger/internal/application/account"
	entryapplication "ledger/internal/application/entry"
	transactionapplication "ledger/internal/application/transaction"
	transferapplication "ledger/internal/application/transfer"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// logger
	logger := logger.New(cfg.Logger.LogLevel, cfg.Logger.LogFormat)

	// connect to postgres
	pgx, err := db.New(cfg.Postgres, logger)
	if err != nil {
		log.Fatal(err)
	}

	// Tx Manager
	txManager := db.NewTxManager(pgx)

	// repositories
	accountRepo := account.NewAccountRepo(pgx)
	transactionRepo := transaction.NewTransactionRepo(pgx)
	entryRepo := entry.NewEntryRepo(pgx)
	idempotencyRepo := idempotency.NewIdempotencyRepo(pgx)

	// applications
	accountApp := accountapplication.NewAccountApplication(accountRepo, logger)
	transferApp := transferapplication.NewTransferApplication(logger, txManager, accountRepo, transactionRepo, entryRepo, idempotencyRepo)
	entryApp := entryapplication.NewEntryApplication(logger, entryRepo)
	transactionApp := transactionapplication.NewTransactionApplication(logger, transactionRepo, entryRepo, txManager, accountRepo)

	// server
	srv := server.New(cfg, logger, pgx, accountApp, transferApp, entryApp, transactionApp)
	if err := srv.Run(); err != nil {
		logger.Error("failed to run server", "error", err)
		log.Fatal(err)
	}
}
