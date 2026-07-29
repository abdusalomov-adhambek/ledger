package outbox_events

import (
	"context"
	"ledger/internal/domain/outbox_events"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxEventRepo struct {
	pxg *pgxpool.Pool
}

func NewOutboxEventRepo(pxg *pgxpool.Pool) *OutboxEventRepo {
	return &OutboxEventRepo{pxg: pxg}
}

func (r *OutboxEventRepo) Create(ctx context.Context, tx pgx.Tx, event *outbox_events.OutboxEvent) error {
	query := `
		INSERT INTO outbox_events (
			event_type,
			aggregate_id,
			payload,
			status
		)
		VALUES ($1, $2, $3, $4)
	`
	_, err := tx.Exec(ctx, query, event.EventType(), event.AggregateId(), event.Payload(), event.Status())
	return err
}

func (r *OutboxEventRepo) GetPending(ctx context.Context) ([]*outbox_events.OutboxEvent, error) {
	query := `
		SELECT
			id,
			event_type,
			aggregate_id,
			payload,
			status,
			created_at
		FROM outbox_events
		WHERE status = 'PENDING'
		LIMIT 100
	`
	rows, err := r.pxg.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*outbox_events.OutboxEvent

	for rows.Next() {
		var (
			id          string
			eventType   string
			aggregateId string
			payload     []byte
			status      string
			createdAt   time.Time
		)

		if err := rows.Scan(&id, &eventType, &aggregateId, &payload, &status, &createdAt); err != nil {
			return nil, err
		}

		newOutBoxEvent := outbox_events.NewOutboxEvent(
			id,
			eventType,
			aggregateId,
			payload,
			status,
		)
		newOutBoxEvent.SetCreatedAt(createdAt)

		events = append(events, newOutBoxEvent)
	}
	return events, nil
}

func (r *OutboxEventRepo) MarkProcessed(ctx context.Context, id string, status string) error {
	query := `
		UPDATE outbox_events
		SET
			status=$2,
			processed_at=NOW()
		WHERE id=$1
	`

	_, err := r.pxg.Exec(ctx, query, id, status)
	return err
}
