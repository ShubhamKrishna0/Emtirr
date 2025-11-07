package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaService struct {
	writer *kafka.Writer
}

func NewKafkaService(brokerURL string) *KafkaService {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokerURL),
		Topic:        "game-analytics",
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	return &KafkaService{
		writer: writer,
	}
}

func (k *KafkaService) PublishEvent(event AnalyticsEvent) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	gameID := ""
	if gid, ok := event.Data["gameId"].(string); ok {
		gameID = gid
	}

	message := kafka.Message{
		Key:   []byte(gameID),
		Value: eventJSON,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return k.writer.WriteMessages(ctx, message)
}

func (k *KafkaService) StartConsumer(processor func(AnalyticsEvent)) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{k.writer.Addr.String()},
		Topic:    "game-analytics",
		GroupID:  "analytics-consumer",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	go func() {
		defer reader.Close()
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			message, err := reader.ReadMessage(ctx)
			cancel()

			if err != nil {
				log.Printf("Kafka consumer error: %v", err)
				continue
			}

			var event AnalyticsEvent
			if err := json.Unmarshal(message.Value, &event); err != nil {
				log.Printf("Failed to unmarshal event: %v", err)
				continue
			}

			processor(event)
		}
	}()
}

func (k *KafkaService) Close() {
	if k.writer != nil {
		k.writer.Close()
	}
}