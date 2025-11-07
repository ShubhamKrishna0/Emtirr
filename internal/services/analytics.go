package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"emitrr-4-in-a-row/internal/config"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type AnalyticsService struct {
	cfg         *config.Config
	kafkaWriter *kafka.Writer
	redisClient *redis.Client
	useRedis    bool
	initialized bool
}

type AnalyticsEvent struct {
	EventType string                 `json:"eventType"`
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

func NewAnalyticsService(cfg *config.Config) *AnalyticsService {
	as := &AnalyticsService{
		cfg: cfg,
	}
	as.Initialize()
	return as
}

func (as *AnalyticsService) Initialize() error {
	brokerURL := as.cfg.KafkaBroker
	if brokerURL == "" && as.cfg.RedisURL == "" {
		log.Println("No broker configured, analytics disabled")
		return nil
	}

	// Try Redis first (for Render)
	if strings.Contains(brokerURL, "redis://") || as.cfg.RedisURL != "" {
		redisURL := as.cfg.RedisURL
		if redisURL == "" {
			redisURL = brokerURL
		}

		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Printf("Redis URL parse error: %v", err)
		} else {
			as.redisClient = redis.NewClient(opt)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := as.redisClient.Ping(ctx).Err(); err == nil {
				as.useRedis = true
				as.initialized = true
				log.Println("Redis Analytics initialized")
				return nil
			} else {
				log.Printf("Redis connection failed: %v", err)
			}
		}
	}

	// Fallback to Kafka
	if brokerURL != "" && !strings.Contains(brokerURL, "redis://") {
		as.kafkaWriter = &kafka.Writer{
			Addr:         kafka.TCP(brokerURL),
			Topic:        "game-events",
			Balancer:     &kafka.LeastBytes{},
			WriteTimeout: 10 * time.Second,
			ReadTimeout:  10 * time.Second,
		}

		as.initialized = true
		log.Println("Kafka Analytics Producer initialized")
		return nil
	}

	log.Println("Analytics initialization failed")
	return fmt.Errorf("no valid analytics backend configured")
}

func (as *AnalyticsService) TrackEvent(eventType string, data map[string]interface{}) {
	event := AnalyticsEvent{
		EventType: eventType,
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      data,
	}

	// Send to Kafka/Redis if available
	if as.initialized {
		if as.useRedis {
			as.publishToRedis(event)
		} else if as.kafkaWriter != nil {
			as.publishToKafka(event)
		}
	}

	// Always log to console for immediate feedback
	log.Printf("Analytics [%s]: gameId=%v, player=%v, timestamp=%s",
		eventType,
		data["gameId"],
		getPlayerFromData(data),
		event.Timestamp,
	)

	// Process specific events for metrics
	switch eventType {
	case "game_started":
		log.Printf("Game started: %v at %s", data["gameId"], event.Timestamp)
	case "game_ended":
		log.Printf("Game ended: %v, Winner: %v, Duration: %v", 
			data["gameId"], data["winner"], data["duration"])
	case "move_made":
		log.Printf("Move tracked: Game %v, Column %v", data["gameId"], data["column"])
	case "bot_move":
		log.Printf("Bot move: Game %v, Column %v", data["gameId"], data["column"])
	}
}

func (as *AnalyticsService) publishToRedis(event AnalyticsEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return
	}

	if err := as.redisClient.LPush(ctx, "game-events", eventJSON).Err(); err != nil {
		log.Printf("Failed to publish to Redis: %v", err)
	}
}

func (as *AnalyticsService) publishToKafka(event AnalyticsEvent) {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return
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

	if err := as.kafkaWriter.WriteMessages(ctx, message); err != nil {
		log.Printf("Failed to send analytics event: %v", err)
	}
}

func (as *AnalyticsService) StartConsumer(dbService *DatabaseService) error {
	if !as.initialized {
		return fmt.Errorf("analytics service not initialized")
	}

	if as.useRedis {
		return as.startRedisConsumer(dbService)
	} else if as.kafkaWriter != nil {
		return as.startKafkaConsumer(dbService)
	}

	return fmt.Errorf("no consumer available")
}

func (as *AnalyticsService) startRedisConsumer(dbService *DatabaseService) error {
	log.Println("Starting Redis analytics consumer")

	go func() {
		ctx := context.Background()
		for {
			result, err := as.redisClient.BRPop(ctx, 1*time.Second, "game-events").Result()
			if err != nil {
				if err != redis.Nil {
					log.Printf("Redis consumer error: %v", err)
				}
				continue
			}

			if len(result) < 2 {
				continue
			}

			var event AnalyticsEvent
			if err := json.Unmarshal([]byte(result[1]), &event); err != nil {
				log.Printf("Failed to unmarshal event: %v", err)
				continue
			}

			as.processEvent(event, dbService)
		}
	}()

	return nil
}

func (as *AnalyticsService) startKafkaConsumer(dbService *DatabaseService) error {
	log.Println("Starting Kafka analytics consumer")

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{as.cfg.KafkaBroker},
		Topic:    "game-events",
		GroupID:  "analytics-group",
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

			as.processEvent(event, dbService)
		}
	}()

	return nil
}

func (as *AnalyticsService) processEvent(event AnalyticsEvent, dbService *DatabaseService) {
	switch event.EventType {
	case "game_started":
		as.trackGameStart(event.Data, event.Timestamp)
	case "game_ended":
		as.trackGameEnd(event.Data, event.Timestamp, dbService)
	case "move_made":
		as.trackMove(event.Data, event.Timestamp)
	default:
		log.Printf("Processing %s event", event.EventType)
	}
}

func (as *AnalyticsService) trackGameStart(data map[string]interface{}, timestamp string) {
	log.Printf("Game started: %v at %s", data["gameId"], timestamp)
}

func (as *AnalyticsService) trackGameEnd(data map[string]interface{}, timestamp string, dbService *DatabaseService) {
	log.Printf("Game ended: %v, Winner: %v, Duration: %v", 
		data["gameId"], data["winner"], data["duration"])

	// Store analytics in database if available
	if dbService != nil {
		gameID := ""
		if gid, ok := data["gameId"].(string); ok {
			gameID = gid
		}

		dbService.SaveAnalyticsEvent("game_analytics", gameID, "", data)
	}
}

func (as *AnalyticsService) trackMove(data map[string]interface{}, timestamp string) {
	log.Printf("Move tracked: Game %v, Column %v", data["gameId"], data["column"])
}

func (as *AnalyticsService) Close() {
	if as.kafkaWriter != nil {
		as.kafkaWriter.Close()
	}
	if as.redisClient != nil {
		as.redisClient.Close()
	}
	log.Println("Analytics service closed")
}

func getPlayerFromData(data map[string]interface{}) interface{} {
	if player, ok := data["player"]; ok {
		return player
	}
	if winner, ok := data["winner"]; ok {
		return winner
	}
	return nil
}