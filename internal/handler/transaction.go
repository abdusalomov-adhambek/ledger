package handler

import (
	"context"
	"ledger/internal/application/transaction"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	logger                 *slog.Logger
	TransactionApplication *transaction.TransactionApplication
}

func NewTransactionHandler(transactionApp *transaction.TransactionApplication, logger *slog.Logger) *TransactionHandler {
	return &TransactionHandler{
		logger:                 logger,
		TransactionApplication: transactionApp,
	}
}

// ReverseTransaction reverses a transaction by its ID
func (h *TransactionHandler) ReverseTransaction(c *gin.Context) {
	h.logger.Info("server.ReverseTransaction", "transaction_id", c.Param("id"))
	var req *transaction.ReverseTransactionRequest
	ctx := context.Background()

	transactionId := c.Param("id")
	if transactionId == "" {
		h.logger.Error("transaction id is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "transaction id is required"})
		return
	}

	req = &transaction.ReverseTransactionRequest{
		TransactionID: transactionId,
	}

	if err := h.TransactionApplication.ReverseTransaction(ctx, req); err != nil {
		h.logger.Error("failed to reverse transaction", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("transaction reversed", "routes.ReverseTransaction", transactionId)
	c.JSON(http.StatusOK, gin.H{"message": "transaction reversed"})
}
