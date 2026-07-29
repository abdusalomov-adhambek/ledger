package handler

import (
	"context"
	"ledger/internal/application/transfer"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TransferHandler struct {
	logger      *slog.Logger
	transferApp *transfer.TransferApp
}

func NewTransferHandler(transferApp *transfer.TransferApp, logger *slog.Logger) *TransferHandler {
	return &TransferHandler{
		logger:      logger,
		transferApp: transferApp,
	}
}

// Transfer transfers funds between accounts
func (h *TransferHandler) Transfer(c *gin.Context) {
	var req transfer.TransferRequest
	ctx := context.Background()

	if err := c.BindJSON(&req); err != nil {
		h.logger.Error("failed to bind request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	idempotencyKey := c.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		h.logger.Info("idempotency key is required")
		c.JSON(http.StatusOK, gin.H{"message": "idempotency key is required"})
		return
	}
	req.IdempotencyKey = idempotencyKey

	transactionId, err := h.transferApp.Transfer(ctx, req)
	if err != nil {
		h.logger.Error("failed to transfer", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("transfer successful")
	c.JSON(http.StatusOK, gin.H{"message": "transfer successful", "id": transactionId})
}
