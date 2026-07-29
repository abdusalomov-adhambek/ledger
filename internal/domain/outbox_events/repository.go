package outbox_events

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Repository interface {
	Create(ctx context.Context, tx pgx.Tx, event *OutboxEvent) error
	GetPending(ctx context.Context) ([]*OutboxEvent, error)
	MarkProcessed(ctx context.Context, id string, status string) error
}
