package entry

import (
	"context"
	"ledger/internal/domain/entry"
	"log/slog"
)

type EntryApplication struct {
	logger *slog.Logger
	repo   entry.Repository
}

func NewEntryApplication(logger *slog.Logger, repo entry.Repository) *EntryApplication {
	return &EntryApplication{
		logger: logger,
		repo:   repo,
	}
}

func (a *EntryApplication) GetHistoryList(ctx context.Context, req *GetHistoryListRequest) ([]*GetListResponse, int64, error) {
	entries, count, err := a.repo.GetList(ctx, &entry.Filter{
		AccountID: &req.AccountID,
	})
	if err != nil {
		a.logger.Error("failed to get history list", err)
		return nil, 0, err
	}

	var data []*GetListResponse

	for _, entry := range entries {
		detail := &GetListResponse{
			ID:            entry.ID(),
			TransactionID: entry.TransactionID(),
			Amount:        entry.Amount(),
			Description:   entry.Transaction().Description(),
		}
		data = append(data, detail)
	}

	a.logger.Info("get list", "count", count)
	return data, count, nil

}
