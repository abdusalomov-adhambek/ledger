package main

import (
	"context"
	"fmt"
	"ledger/internal/adapter/postgres/account"
	"ledger/internal/adapter/postgres/entry"
	"ledger/internal/adapter/postgres/idempotency"
	"ledger/internal/adapter/postgres/outbox_events"
	"ledger/internal/adapter/postgres/transaction"
	"ledger/internal/config"
	"ledger/internal/db"
	"ledger/internal/logger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	accountapplication "ledger/internal/application/account"
	entryapplication "ledger/internal/application/entry"
	transactionapplication "ledger/internal/application/transaction"
	transferapplication "ledger/internal/application/transfer"

	handler "ledger/internal/handler"

	"ledger/internal/worker"

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
	outboxEventRepo := outbox_events.NewOutboxEventRepo(pgx)

	// applications
	accountApp := accountapplication.NewAccountApplication(accountRepo, logger)
	transferApp := transferapplication.NewTransferApplication(logger, txManager, accountRepo, transactionRepo, entryRepo, idempotencyRepo, outboxEventRepo)
	entryApp := entryapplication.NewEntryApplication(logger, entryRepo)
	transactionApp := transactionapplication.NewTransactionApplication(logger, transactionRepo, entryRepo, txManager, accountRepo)

	// handlers
	accountHandler := handler.NewAccountHandler(accountApp, logger)
	transactionHandler := handler.NewTransactionHandler(transactionApp, logger)
	transferHandler := handler.NewTransferHandler(transferApp, logger)
	entryHandler := handler.NewEntryHandler(entryApp, logger)
	healthHandler := handler.NewHealthHandler()

	// outbox events worker
	worker := worker.NewOutboxWorker(outboxEventRepo, logger)

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

	// app.Run(fmt.Sprintf(":%d", cfg.Port))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: app,
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	go func() {
		<-ctx.Done()

		logger.Info("shutting down server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("failed to shutdown server", "error", err)
		}

	}()

	go worker.Run(ctx)

	logger.Info("server started", "port", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}
