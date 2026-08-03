package main

import (
	"context"
	"fmt"
	"ledger/internal/config"
	"ledger/internal/logger"
	"log"

	"github.com/segmentio/kafka-go"
)

func main() {
	// config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// logger
	logger := logger.New(cfg.Logger.LogLevel, cfg.Logger.LogFormat)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "leader.events",
		GroupID: "leader-consumer",
	})
	defer reader.Close()

	logger.Info("Kafka consumer started...")

	ctx := context.Background()

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("---------------")
		fmt.Println("Topic:", msg.Topic)
		fmt.Println("Key:", string(msg.Key))
		fmt.Println("Value:", string(msg.Value))
		fmt.Println("Offset:", msg.Offset)
		fmt.Println("---------------")

	}
}
