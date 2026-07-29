package idempotency

import "time"

type Idempotency struct {
	key         string
	trsactionId string
	createdAt   *time.Time
}

func NewIdempotency(
	key string,
	transactionId string,
	createdAt *time.Time,
) *Idempotency {
	return &Idempotency{
		key:         key,
		trsactionId: transactionId,
		createdAt:   createdAt,
	}
}

func (i *Idempotency) Key() string {
	return i.key
}

func (i *Idempotency) SetKey(key string) {
	i.key = key
}

func (i *Idempotency) TransactionID() string {
	return i.trsactionId
}

func (i *Idempotency) SetTransactionID(transactionId string) {
	i.trsactionId = transactionId
}

func (i *Idempotency) CreatedAt() *time.Time {
	return i.createdAt
}
