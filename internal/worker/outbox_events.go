package worker

import (
	"context"
	"ledger/internal/domain/outbox_events"
	"log/slog"
	"time"
)

type OutboxWorker struct {
	outboxRepo outbox_events.Repository
	logger     *slog.Logger
}

func NewOutboxWorker(outboxRepo outbox_events.Repository, logger *slog.Logger) *OutboxWorker {
	return &OutboxWorker{
		outboxRepo: outboxRepo,
		logger:     logger,
	}
}

func (w *OutboxWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("outbox worker stopped")
			return
		case <-ticker.C:
			w.process(ctx)
		}
	}
}

func (w *OutboxWorker) process(ctx context.Context) {
	events, err := w.outboxRepo.GetPending(ctx)
	if err != nil {
		w.logger.Error("failed to get pending events", "error", err)
		return
	}

	for _, event := range events {
		if err := w.publish(ctx, event); err != nil {
			continue
		}

		if err := w.outboxRepo.MarkProcessed(ctx, event.ID(), outbox_events.StatusProcessed); err != nil {
			w.logger.Error(
				"mark processed",
				"id", event.ID(),
				"error", err,
			)
		}
	}
}

func (w *OutboxWorker) publish(ctx context.Context, event *outbox_events.OutboxEvent) error {
	w.logger.Info("publishing event", "id", event.ID())
	return nil
}
