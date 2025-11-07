package services

import (
	"encoding/json"
	"log"
	"time"

	"github.com/IBM/sarama"
)

type KafkaService struct {
	producer sarama.SyncProducer
	topic    string
}

type AnalyticsEvent struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

func NewKafkaService(brokers []string, topic string) (*KafkaService, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		// If Kafka is not available, return nil service (graceful degradation)
		log.Printf("Kafka not available, analytics will be disabled: %v", err)
		return nil, nil
	}

	return &KafkaService{
		producer: producer,
		topic:    topic,
	}, nil
}

func (k *KafkaService) PublishEvent(eventType string, data map[string]interface{}) {
	if k == nil || k.producer == nil {
		return // Graceful degradation when Kafka is not available
	}

	event := AnalyticsEvent{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal analytics event: %v", err)
		return
	}

	message := &sarama.ProducerMessage{
		Topic: k.topic,
		Value: sarama.StringEncoder(eventBytes),
	}

	_, _, err = k.producer.SendMessage(message)
	if err != nil {
		log.Printf("Failed to send analytics event: %v", err)
	}
}

func (k *KafkaService) Close() {
	if k != nil && k.producer != nil {
		k.producer.Close()
	}
}