package kafka

import (
	"context"
	"ledger/internal/domain/outbox_events"

	"github.com/segmentio/kafka-go"
)

type Producer interface {
	Publish(ctx context.Context, event *outbox_events.OutboxEvent) error
}

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *KafkaProducer) Publish(ctx context.Context, event *outbox_events.OutboxEvent) error {
	return p.writer.WriteMessages(
		ctx,
		kafka.Message{
			Key:   []byte(event.AggregateID()),
			Value: event.Payload(),
		},
	)
}
