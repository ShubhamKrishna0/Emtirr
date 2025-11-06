package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"emitrr-4-in-a-row/internal/config"

	_ "github.com/lib/pq"
)

type DatabaseService struct {
	db  *sql.DB
	cfg *config.Config
}

type GameData struct {
	ID        string
	Player1   string
	Player2   string
	Winner    *int
	Duration  int
	Moves     int
	IsBot     bool
	CreatedAt time.Time
}

type PlayerStats struct {
	Username    string  `json:"username"`
	GamesPlayed int     `json:"games_played"`
	GamesWon    int     `json:"games_won"`
	WinRate     float64 `json:"win_rate"`
	LastPlayed  string  `json:"last_played"`
}

type Analytics struct {
	TotalGames      []map[string]interface{} `json:"totalGames"`
	TotalPlayers    []map[string]interface{} `json:"totalPlayers"`
	AvgGameDuration []map[string]interface{} `json:"avgGameDuration"`
	GamesPerDay     []map[string]interface{} `json:"gamesPerDay"`
	TopWinners      []map[string]interface{} `json:"topWinners"`
	BotVsHuman      []map[string]interface{} `json:"botVsHuman"`
}

func NewDatabaseService(cfg *config.Config) *DatabaseService {
	return &DatabaseService{cfg: cfg}
}

func (ds *DatabaseService) Initialize() error {
	var connStr string
	if ds.cfg.DatabaseURL != "" {
		connStr = ds.cfg.DatabaseURL
	} else {
		connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			ds.cfg.DBHost, ds.cfg.DBPort, ds.cfg.DBUser, ds.cfg.DBPassword, ds.cfg.DBName)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	ds.db = db
	return ds.createTables()
}

func (ds *DatabaseService) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS games (
			id VARCHAR(36) PRIMARY KEY,
			player1 VARCHAR(100) NOT NULL,
			player2 VARCHAR(100) NOT NULL,
			winner INTEGER,
			duration INTEGER NOT NULL,
			moves INTEGER NOT NULL,
			is_bot BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			finished_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS players (
			username VARCHAR(100) PRIMARY KEY,
			games_played INTEGER DEFAULT 0,
			games_won INTEGER DEFAULT 0,
			total_duration INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_played TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS analytics_events (
			id SERIAL PRIMARY KEY,
			event_type VARCHAR(50) NOT NULL,
			game_id VARCHAR(36),
			player VARCHAR(100),
			data JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_games_created_at ON games(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_players_games_won ON players(games_won DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_events_type ON analytics_events(event_type)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_events_created_at ON analytics_events(created_at)`,
	}

	for _, query := range queries {
		if _, err := ds.db.Exec(query); err != nil {
			return err
		}
	}

	log.Println("Database tables created successfully")
	return nil
}

func (ds *DatabaseService) SaveGame(gameData GameData) error {
	if ds.db == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `
		INSERT INTO games (id, player1, player2, winner, duration, moves, is_bot, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := ds.db.Exec(query,
		gameData.ID,
		gameData.Player1,
		gameData.Player2,
		gameData.Winner,
		gameData.Duration,
		gameData.Moves,
		gameData.IsBot,
		gameData.CreatedAt,
	)

	return err
}

func (ds *DatabaseService) UpdatePlayerStats(username string, won bool) error {
	if ds.db == nil {
		return fmt.Errorf("database not initialized")
	}

	wonInt := 0
	if won {
		wonInt = 1
	}

	query := `
		INSERT INTO players (username, games_played, games_won, last_played)
		VALUES ($1, 1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (username) 
		DO UPDATE SET 
			games_played = players.games_played + 1,
			games_won = players.games_won + $2,
			last_played = CURRENT_TIMESTAMP
	`

	_, err := ds.db.Exec(query, username, wonInt)
	return err
}

func (ds *DatabaseService) GetLeaderboard(limit int) ([]PlayerStats, error) {
	if ds.db == nil {
		return []PlayerStats{}, nil
	}

	query := `
		SELECT 
			username,
			games_played,
			games_won,
			ROUND((games_won::DECIMAL / GREATEST(games_played, 1)) * 100, 1) as win_rate,
			last_played
		FROM players 
		WHERE games_played > 0
		ORDER BY games_won DESC, win_rate DESC, games_played DESC
		LIMIT $1
	`

	rows, err := ds.db.Query(query, limit)
	if err != nil {
		return []PlayerStats{}, err
	}
	defer rows.Close()

	var leaderboard []PlayerStats
	for rows.Next() {
		var stats PlayerStats
		var lastPlayed time.Time
		err := rows.Scan(&stats.Username, &stats.GamesPlayed, &stats.GamesWon, &stats.WinRate, &lastPlayed)
		if err != nil {
			continue
		}
		stats.LastPlayed = lastPlayed.Format("2006-01-02 15:04:05")
		leaderboard = append(leaderboard, stats)
	}

	return leaderboard, nil
}

func (ds *DatabaseService) GetAnalytics() (Analytics, error) {
	analytics := Analytics{}

	if ds.db == nil {
		return analytics, nil
	}

	queries := map[string]string{
		"totalGames":      "SELECT COUNT(*) as count FROM games",
		"totalPlayers":    "SELECT COUNT(*) as count FROM players WHERE games_played > 0",
		"avgGameDuration": "SELECT ROUND(AVG(duration), 1) as avg_duration FROM games",
		"gamesPerDay": `
			SELECT 
				DATE(created_at) as date,
				COUNT(*) as games
			FROM games 
			WHERE created_at >= CURRENT_DATE - INTERVAL '7 days'
			GROUP BY DATE(created_at)
			ORDER BY date DESC
		`,
		"topWinners": `
			SELECT username, games_won 
			FROM players 
			WHERE games_played > 0
			ORDER BY games_won DESC 
			LIMIT 5
		`,
		"botVsHuman": `
			SELECT 
				is_bot,
				COUNT(*) as count,
				ROUND(AVG(duration), 1) as avg_duration
			FROM games 
			GROUP BY is_bot
		`,
	}

	// Total Games
	if rows, err := ds.db.Query(queries["totalGames"]); err == nil {
		defer rows.Close()
		for rows.Next() {
			var count int
			rows.Scan(&count)
			analytics.TotalGames = append(analytics.TotalGames, map[string]interface{}{"count": count})
		}
	}

	// Total Players
	if rows, err := ds.db.Query(queries["totalPlayers"]); err == nil {
		defer rows.Close()
		for rows.Next() {
			var count int
			rows.Scan(&count)
			analytics.TotalPlayers = append(analytics.TotalPlayers, map[string]interface{}{"count": count})
		}
	}

	// Average Game Duration
	if rows, err := ds.db.Query(queries["avgGameDuration"]); err == nil {
		defer rows.Close()
		for rows.Next() {
			var avgDuration sql.NullFloat64
			rows.Scan(&avgDuration)
			if avgDuration.Valid {
				analytics.AvgGameDuration = append(analytics.AvgGameDuration, map[string]interface{}{"avg_duration": avgDuration.Float64})
			}
		}
	}

	// Games Per Day
	if rows, err := ds.db.Query(queries["gamesPerDay"]); err == nil {
		defer rows.Close()
		for rows.Next() {
			var date time.Time
			var games int
			rows.Scan(&date, &games)
			analytics.GamesPerDay = append(analytics.GamesPerDay, map[string]interface{}{
				"date":  date.Format("2006-01-02"),
				"games": games,
			})
		}
	}

	// Top Winners
	if rows, err := ds.db.Query(queries["topWinners"]); err == nil {
		defer rows.Close()
		for rows.Next() {
			var username string
			var gamesWon int
			rows.Scan(&username, &gamesWon)
			analytics.TopWinners = append(analytics.TopWinners, map[string]interface{}{
				"username":  username,
				"games_won": gamesWon,
			})
		}
	}

	// Bot vs Human
	if rows, err := ds.db.Query(queries["botVsHuman"]); err == nil {
		defer rows.Close()
		for rows.Next() {
			var isBot bool
			var count int
			var avgDuration sql.NullFloat64
			rows.Scan(&isBot, &count, &avgDuration)
			analytics.BotVsHuman = append(analytics.BotVsHuman, map[string]interface{}{
				"is_bot":       isBot,
				"count":        count,
				"avg_duration": avgDuration.Float64,
			})
		}
	}

	return analytics, nil
}

func (ds *DatabaseService) SaveAnalyticsEvent(eventType, gameID, player string, data map[string]interface{}) error {
	if ds.db == nil {
		return nil
	}

	query := `
		INSERT INTO analytics_events (event_type, game_id, player, data)
		VALUES ($1, $2, $3, $4)
	`

	dataJSON := "{}"
	if data != nil {
		// Simple JSON marshaling - in production, use proper JSON library
		dataJSON = fmt.Sprintf("%v", data)
	}

	_, err := ds.db.Exec(query, eventType, gameID, player, dataJSON)
	return err
}

func (ds *DatabaseService) Close() error {
	if ds.db != nil {
		return ds.db.Close()
	}
	return nil
}