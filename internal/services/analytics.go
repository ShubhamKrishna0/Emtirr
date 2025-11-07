package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"emitrr-4-in-a-row/internal/config"

	"github.com/redis/go-redis/v9"
)

type AnalyticsService struct {
	cfg         *config.Config
	redisClient *redis.Client
	kafkaService *KafkaService
	useKafka    bool
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
	// Try Kafka first (for local demo)
	if as.cfg.KafkaBroker != "" {
		as.kafkaService = NewKafkaService(as.cfg.KafkaBroker)
		as.useKafka = true
		as.initialized = true
		log.Println("Kafka Analytics initialized for demo")
		return nil
	}

	// Fallback to Redis (for production)
	if as.cfg.RedisURL != "" {
		opt, err := redis.ParseURL(as.cfg.RedisURL)
		if err != nil {
			log.Printf("Redis URL parse error: %v", err)
			return err
		}

		as.redisClient = redis.NewClient(opt)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := as.redisClient.Ping(ctx).Err(); err == nil {
			as.initialized = true
			log.Println("Redis Analytics initialized")
			return nil
		} else {
			log.Printf("Redis connection failed: %v", err)
		}
	}

	log.Println("No analytics backend configured")
	return nil
}

func (as *AnalyticsService) TrackEvent(eventType string, data map[string]interface{}) {
	event := AnalyticsEvent{
		EventType: eventType,
		Timestamp: time.Now().Format(time.RFC3339),
		Data:      data,
	}

	if as.initialized {
		if as.useKafka {
			if err := as.kafkaService.PublishEvent(event); err != nil {
				log.Printf("Failed to publish to Kafka: %v", err)
			}
		} else if as.redisClient != nil {
			as.publishToRedis(event)
		}
	}

	log.Printf("Analytics [%s]: gameId=%v, player=%v, timestamp=%s",
		eventType,
		data["gameId"],
		getPlayerFromData(data),
		event.Timestamp,
	)

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

func (as *AnalyticsService) StartConsumer(dbService *DatabaseService) error {
	if !as.initialized {
		log.Println("Analytics service not initialized, skipping consumer")
		return nil
	}

	if as.useKafka {
		as.kafkaService.StartConsumer(func(event AnalyticsEvent) {
			as.processEvent(event, dbService)
		})
		log.Println("Kafka consumer started for demo")
		return nil
	}

	return as.startRedisConsumer(dbService)
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
	if as.kafkaService != nil {
		as.kafkaService.Close()
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