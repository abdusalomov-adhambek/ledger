package transaction

import "encoding/json"

type Transaction struct {
	id          string
	reference   string
	status      string
	description string
	metadata    json.RawMessage
	createdAt   string
}

const (
	StatusPending   string = "PENDING"
	StatusCompleted string = "COMPLETED"
	StatusCanceled  string = "CANCELLED"
	StatusReversed  string = "REVERSED"
)

func NewTransaction(
	id string,
	reference string,
	status string,
	description string,
	metadata json.RawMessage,
) *Transaction {
	return &Transaction{
		id:          id,
		reference:   reference,
		status:      status,
		description: description,
		metadata:    metadata,
	}
}

func (t *Transaction) ID() string {
	return t.id
}

func (t *Transaction) SetID(id string) {
	t.id = id
}

func (t *Transaction) Reference() string {
	return t.reference
}

func (t *Transaction) Status() string {
	return t.status
}

func (t *Transaction) Description() string {
	return t.description
}

func (t *Transaction) Metadata() json.RawMessage {
	return t.metadata
}

func (t *Transaction) CreatedAt() string {
	return t.createdAt
}
