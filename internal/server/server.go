package server

import (
	"fmt"
	"ledger/internal/application/account"
	"ledger/internal/application/entry"
	"ledger/internal/application/transaction"
	"ledger/internal/application/transfer"
	"ledger/internal/config"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	router *gin.Engine
	cfg    *config.Config
	logger *slog.Logger
	db     *pgxpool.Pool

	accountApp     *account.AccountApplication
	transferApp    *transfer.TransferApp
	entryApp       *entry.EntryApplication
	transactionApp *transaction.TransactionApplication
}

func New(
	cfg *config.Config,
	logger *slog.Logger,
	db *pgxpool.Pool,
	accountApp *account.AccountApplication,
	transferApp *transfer.TransferApp,
	entryApp *entry.EntryApplication,
	transactionApp *transaction.TransactionApplication,
) *Server {
	s := &Server{
		router:         gin.New(),
		cfg:            cfg,
		logger:         logger,
		db:             db,
		accountApp:     accountApp,
		transferApp:    transferApp,
		entryApp:       entryApp,
		transactionApp: transactionApp,
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	s.router.GET("/", s.health)

	// ------------------- account -------------------
	s.router.POST("/account", s.CreateAccount)
	s.router.GET("/account/:id", s.GetByID)

	// ------------------- transfer -------------------
	s.router.POST("/transfer", s.Transfer)

	// ------------------- entry -------------------
	s.router.GET("/entry/:account_id/history", s.GetHistoryEntries)

	// ------------------- transaction -------------------
	s.router.PUT("/transaction/:id/reverse", s.ReverseTransaction)
}

func (s *Server) Run() error {
	s.logger.Info("server is running", "port", s.cfg.Port)
	return s.router.Run(fmt.Sprintf(":%d", s.cfg.Port))
}
