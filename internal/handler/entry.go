package handler

import (
	"context"
	"ledger/internal/application/entry"
	"log/slog"
	"net/http"
	"strconv"

	entrydomain "ledger/internal/domain/entry"

	"github.com/gin-gonic/gin"
)

type EntryHandler struct {
	logger           *slog.Logger
	entryApplication *entry.EntryApplication
}

func NewEntryHandler(entryApplication *entry.EntryApplication, logger *slog.Logger) *EntryHandler {
	return &EntryHandler{
		logger:           logger,
		entryApplication: entryApplication,
	}
}

// GetEntries retrieves a list of entries
func (h *EntryHandler) GetHistoryEntries(c *gin.Context) {
	req := &entry.GetHistoryListRequest{}
	filter := &entrydomain.Filter{}

	accountId := c.Param("account_id")
	if accountId == "" {
		h.logger.Error("account_id is required")
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

	entries, count, err := h.entryApplication.GetHistoryList(context.Background(), filter, req)

	if err != nil {
		h.logger.Error("failed to get entries", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("entries retrieved", "count", count)
	c.JSON(http.StatusOK, gin.H{
		"results": gin.H{
			"data":  entries,
			"count": count,
		},
	})
}
