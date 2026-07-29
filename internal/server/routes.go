package server

import (
	"context"
	"ledger/internal/application/account"
	"ledger/internal/application/entry"
	"ledger/internal/application/transaction"
	"ledger/internal/application/transfer"
	"strconv"

	entrydomain "ledger/internal/domain/entry"

	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// CreateAccount creates a new account
func (s *Server) CreateAccount(c *gin.Context) {
	var req account.CreateAccountRequest
	ctx := context.Background()

	if err := c.BindJSON(&req); err != nil {
		s.logger.Error("failed to bind request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	detail, err := s.accountApp.CreateAccount(ctx, req)
	if err != nil {
		s.logger.Error("failed to create account", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.logger.Info("account created", "id", detail.ID)
	c.JSON(http.StatusOK, gin.H{
		"id": detail.ID,
	})
}

// GetByID retrieves an account by its ID
func (s *Server) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		s.logger.Error("id is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	detail, err := s.accountApp.GetByID(context.Background(), account.GetByIDRequest{
		ID: id,
	})
	if err != nil {
		s.logger.Error("failed to get account", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.logger.Info("account retrieved", "id", id)
	c.JSON(http.StatusOK, gin.H{
		"result": detail,
	})
}

// Transfer transfers funds between accounts
func (s *Server) Transfer(c *gin.Context) {
	var req transfer.TransferRequest
	ctx := context.Background()

	if err := c.BindJSON(&req); err != nil {
		s.logger.Error("failed to bind request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	idempotencyKey := c.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		s.logger.Info("idempotency key is required")
		c.JSON(http.StatusOK, gin.H{"message": "idempotency key is required"})
		return
	}
	req.IdempotencyKey = idempotencyKey

	transactionId, err := s.transferApp.Transfer(ctx, req)
	if err != nil {
		s.logger.Error("failed to transfer", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.logger.Info("transfer successful")
	c.JSON(http.StatusOK, gin.H{"message": "transfer successful", "id": transactionId})
}

// GetEntries retrieves a list of entries
func (s *Server) GetHistoryEntries(c *gin.Context) {
	req := &entry.GetHistoryListRequest{}
	filter := &entrydomain.Filter{}

	accountId := c.Param("account_id")
	if accountId == "" {
		s.logger.Error("account_id is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "account_id is required"})
		return
	}
	req.AccountID = accountId

	if limit := c.Query("limit"); limit != "" {
		l, err := strconv.ParseInt(limit, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
			return
		}
		filter.Limit = &l
	}

	if offset := c.Query("offset"); offset != "" {
		o, err := strconv.ParseInt(offset, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})
			return
		}
		filter.Offset = &o
	}

	entries, count, err := s.entryApp.GetHistoryList(context.Background(), filter, req)

	if err != nil {
		s.logger.Error("failed to get entries", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.logger.Info("entries retrieved", "count", count)
	c.JSON(http.StatusOK, gin.H{
		"results": gin.H{
			"data":  entries,
			"count": count,
		},
	})
}

// ReverseTransaction reverses a transaction by its ID
func (s *Server) ReverseTransaction(c *gin.Context) {
	s.logger.Info("server.ReverseTransaction", "transaction_id", c.Param("id"))
	var req *transaction.ReverseTransactionRequest
	ctx := context.Background()

	transactionId := c.Param("id")
	if transactionId == "" {
		s.logger.Error("transaction id is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "transaction id is required"})
		return
	}

	req = &transaction.ReverseTransactionRequest{
		TransactionID: transactionId,
	}

	if err := s.transactionApp.ReverseTransaction(ctx, req); err != nil {
		s.logger.Error("failed to reverse transaction", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.logger.Info("transaction reversed", "routes.ReverseTransaction", transactionId)
	c.JSON(http.StatusOK, gin.H{"message": "transaction reversed"})
}
