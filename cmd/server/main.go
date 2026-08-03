package main

import (
	"context"
	"fmt"
	ledgerv1 "ledger/api/proto"
	"ledger/internal/adapter/kafka"
	"ledger/internal/adapter/postgres/account"
	"ledger/internal/adapter/postgres/entry"
	"ledger/internal/adapter/postgres/idempotency"
	"ledger/internal/adapter/postgres/outbox_events"
	"ledger/internal/adapter/postgres/transaction"
	"ledger/internal/config"
	"ledger/internal/db"
	"ledger/internal/logger"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	accountapplication "ledger/internal/application/account"
	entryapplication "ledger/internal/application/entry"
	transactionapplication "ledger/internal/application/transaction"
	transferapplication "ledger/internal/application/transfer"

	ledgergrpc "ledger/internal/adapter/grpc"

	"ledger/internal/worker"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	// kafka
	kafkaProducer := kafka.NewKafkaProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)

	// ctx
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

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

	// outbox events worker
	worker := worker.NewOutboxWorker(outboxEventRepo, logger, kafkaProducer)
	logger.Info("outbox worker started", worker)

	//  --------------------------- gRPC gateway ---------------------------
	gatewayMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(headerMatcher),
	)

	if err := ledgerv1.RegisterLedgerServiceHandlerFromEndpoint(
		ctx,
		gatewayMux,
		fmt.Sprintf("localhost:%s", cfg.GRPC.LedgerPort),
		[]grpc.DialOption{
			grpc.WithTransportCredentials(
				insecure.NewCredentials(),
			),
		},
	); err != nil {
		return err
	}

	gatewayServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: gatewayMux,
	}

	// --------------------------- gRPC server ---------------------------
	grpcServer := grpc.NewServer()

	ledgerGrpcService := ledgergrpc.NewLedgerGRPCService(
		accountApp,
		transferApp,
		transactionApp,
		entryApp)

	ledgerv1.RegisterLedgerServiceServer(
		grpcServer, ledgerGrpcService,
	)

	lis, err := net.Listen("tcp", ":"+cfg.GRPC.LedgerPort)
	if err != nil {
		return err
	}

	go func() {
		logger.Info("gRPC server is listening", "port", cfg.GRPC.LedgerPort)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server failed", "error", err)
		}
	}()

	// --------------------------- Shutdown ---------------------------
	go func() {
		<-ctx.Done()

		logger.Info("shutting down server")

		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := gatewayServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("failed to shutdown server", "error", err)
		}

		grpcServer.GracefulStop()
		logger.Info("grpc server stopped")
	}()

	// --------------------------- Worker run ---------------------------
	// go worker.Run(ctx)

	// --------------------------- gRPC Gateway run ---------------------------
	logger.Info("grpc gateway started", "port", cfg.Port)
	if err := gatewayServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("gateway stopped", "error", err)
	}

	return nil
}

func headerMatcher(key string) (string, bool) {
	switch key {
	case "Idempotency-Key":
		return "idempotency-key", true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}
