package main

import (
	"fmt"
	"ledger/internal/adapter/postgres/account"
	"ledger/internal/adapter/postgres/entry"
	"ledger/internal/adapter/postgres/idempotency"
	"ledger/internal/adapter/postgres/transaction"
	"ledger/internal/config"
	"ledger/internal/db"
	"ledger/internal/logger"
	"log"

	accountapplication "ledger/internal/application/account"
	entryapplication "ledger/internal/application/entry"
	transactionapplication "ledger/internal/application/transaction"
	transferapplication "ledger/internal/application/transfer"

	handler "ledger/internal/handler"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}

func run() error {
	// config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
		return err
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

	// handlers
	accountHandler := handler.NewAccountHandler(accountApp, logger)
	transactionHandler := handler.NewTransactionHandler(transactionApp, logger)
	transferHandler := handler.NewTransferHandler(transferApp, logger)
	entryHandler := handler.NewEntryHandler(entryApp, logger)
	healthHandler := handler.NewHealthHandler()

	app := gin.Default()

	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	// ------------------ health ------------------
	app.GET("/", healthHandler.Health)

	// ------------------- account -------------------
	app.POST("/account", accountHandler.CreateAccount)
	app.GET("/account/:id", accountHandler.GetByID)

	// ------------------- transfer -------------------
	app.POST("/transfer", transferHandler.Transfer)

	// ------------------- entry -------------------
	app.GET("/entry/:account_id/history", entryHandler.GetHistoryEntries)

	// ------------------- transaction -------------------
	app.PUT("/transaction/:id/reverse", transactionHandler.ReverseTransaction)

	logger.Info("server started", "port", cfg.Port)
	app.Run(fmt.Sprintf(":%d", cfg.Port))
	return nil
}
