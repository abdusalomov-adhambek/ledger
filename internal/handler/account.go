package handler

import (
	"context"
	"ledger/internal/application/account"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	accountApplication *account.AccountApplication
	logger             *slog.Logger
}

func NewAccountHandler(accountApplication *account.AccountApplication, logger *slog.Logger) *AccountHandler {
	return &AccountHandler{
		accountApplication: accountApplication,
		logger:             logger,
	}
}

// CreateAccount creates a new account
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var req account.CreateAccountRequest
	ctx := context.Background()

	if err := c.BindJSON(&req); err != nil {
		h.logger.Error("failed to bind request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	detail, err := h.accountApplication.CreateAccount(ctx, req)
	if err != nil {
		h.logger.Error("failed to create account", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("account created", "id", detail.ID)
	c.JSON(http.StatusOK, gin.H{
		"id": detail.ID,
	})
}

// GetByID retrieves an account by its ID
func (h *AccountHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.logger.Error("id is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	detail, err := h.accountApplication.GetByID(context.Background(), account.GetByIDRequest{
		ID: id,
	})
	if err != nil {
		h.logger.Error("failed to get account", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("account retrieved", "id", id)
	c.JSON(http.StatusOK, gin.H{
		"result": detail,
	})
}
