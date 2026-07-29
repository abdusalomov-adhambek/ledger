package outbox_events

import "time"

type OutboxEvent struct {
	id          string
	eventType   string
	aggregateId string
	payload     []byte
	status      string
	createdAt   *time.Time
	processedAt *time.Time
}

const (
	EventTransferCompleted string = "TRANSFER_COMPLETED"
)

const (
	StatusPending    string = "PENDING"
	StatusProcessing string = "PROCESSING"
	StatusProcessed  string = "PROCESSED"
	StatusFailed     string = "FAILED"
)

func NewOutboxEvent(id, eventType, aggregateId string, payload []byte, status string) *OutboxEvent {
	return &OutboxEvent{
		id:          id,
		eventType:   eventType,
		aggregateId: aggregateId,
		payload:     payload,
		status:      status,
	}
}

func (e *OutboxEvent) ID() string {
	return e.id
}

func (e *OutboxEvent) SetID(id string) {
	e.id = id
}

func (e *OutboxEvent) EventType() string {
	return e.eventType
}

func (e *OutboxEvent) AggregateId() string {
	return e.aggregateId
}

func (e *OutboxEvent) Payload() []byte {
	return e.payload
}

func (e *OutboxEvent) Status() string {
	return e.status
}

func (e *OutboxEvent) CreatedAt() *time.Time {
	return e.createdAt
}

func (e *OutboxEvent) SetCreatedAt(createdAt time.Time) {
	e.createdAt = &createdAt
}

func (e *OutboxEvent) SetProcessedAt(processedAt time.Time) {
	e.processedAt = &processedAt
}

func (e *OutboxEvent) ProcessedAt() *time.Time {
	return e.processedAt
}
