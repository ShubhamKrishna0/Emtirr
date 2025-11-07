package services

import (
	"log"
	"time"
)

type AnalyticsService struct {
	kafkaService *KafkaService
	dbService    *DatabaseService
}

func NewAnalyticsService(kafkaService *KafkaService, dbService *DatabaseService) *AnalyticsService {
	return &AnalyticsService{
		kafkaService: kafkaService,
		dbService:    dbService,
	}
}

func (a *AnalyticsService) TrackEvent(eventType string, data map[string]interface{}) {
	// Add timestamp and session info
	enrichedData := make(map[string]interface{})
	for k, v := range data {
		enrichedData[k] = v
	}
	enrichedData["timestamp"] = time.Now()
	enrichedData["server_id"] = "game-server-1"

	// Send to Kafka for real-time processing
	if a.kafkaService != nil {
		a.kafkaService.PublishEvent(eventType, enrichedData)
	}

	// Also log locally for immediate visibility
	log.Printf("Analytics [%s]: %+v", eventType, enrichedData)

	// Store critical events in database for persistence
	if a.shouldPersist(eventType) {
		a.persistEvent(eventType, enrichedData)
	}
}

func (a *AnalyticsService) shouldPersist(eventType string) bool {
	persistentEvents := map[string]bool{
		"game_started": true,
		"game_ended":   true,
		"player_won":   true,
	}
	return persistentEvents[eventType]
}

func (a *AnalyticsService) persistEvent(eventType string, data map[string]interface{}) {
	if a.dbService == nil {
		return
	}

	// Store in analytics_events table for later analysis
	go func() {
		query := `
			INSERT INTO analytics_events (event_type, event_data, created_at) 
			VALUES ($1, $2, $3)
		`
		
		dataJSON, _ := json.Marshal(data)
		_, err := a.dbService.db.Exec(query, eventType, string(dataJSON), time.Now())
		if err != nil {
			log.Printf("Failed to persist analytics event: %v", err)
		}
	}()
}

func (a *AnalyticsService) Close() {
	if a.kafkaService != nil {
		a.kafkaService.Close()
	}
}