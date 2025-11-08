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
	enabled bool
}

func NewKafkaService(brokerURL string) *KafkaService {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokerURL),
		Topic:        "game-analytics",
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	testMsg := kafka.Message{Key: []byte("test"), Value: []byte("test")}
	err := writer.WriteMessages(ctx, testMsg)
	enabled := err == nil
	
	if !enabled {
		log.Printf("Kafka unavailable, analytics disabled: %v", err)
	}

	return &KafkaService{
		writer: writer,
		enabled: enabled,
	}
}

func (k *KafkaService) PublishEvent(event AnalyticsEvent) error {
	if !k.enabled {
		return nil // Silently skip if Kafka unavailable
	}
	
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = k.writer.WriteMessages(ctx, message)
	if err != nil {
		k.enabled = false // Disable on error
	}
	return err
}

func (k *KafkaService) StartConsumer(processor func(AnalyticsEvent)) {
	if !k.enabled {
		log.Println("Kafka consumer disabled")
		return
	}
	
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{k.writer.Addr.String()},
		Topic:    "game-analytics",
		GroupID:  "analytics-consumer",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	log.Println("Kafka consumer started")
	go func() {
		defer reader.Close()
		errorCount := 0
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			message, err := reader.ReadMessage(ctx)
			cancel()

			if err != nil {
				errorCount++
				if errorCount <= 3 {
					log.Printf("Kafka consumer error (%d/3): %v", errorCount, err)
				}
				if errorCount >= 10 {
					log.Println("Too many Kafka errors, stopping consumer")
					return
				}
				time.Sleep(time.Duration(errorCount) * time.Second)
				continue
			}

			errorCount = 0
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