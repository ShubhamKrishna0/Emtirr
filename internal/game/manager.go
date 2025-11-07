package game

import (
	"log"
	"strings"
	"sync"
	"time"

	"emitrr-4-in-a-row/internal/models"
	"emitrr-4-in-a-row/internal/services"

	"github.com/gorilla/websocket"
)

type GameManager struct {
	games            map[string]*models.Game
	connections      map[*websocket.Conn]*Player
	waitingQueue     []*Player
	disconnected     map[string]*DisconnectedInfo
	dbService        *services.DatabaseService
	analyticsService *services.AnalyticsService
	bot              *Bot
	mu               sync.RWMutex
}

type Player struct {
	Username  string
	Conn      *websocket.Conn
	GameID    string
	PlayerNum int
}

type DisconnectedInfo struct {
	GameID    string
	PlayerNum int
	Time      time.Time
}

func NewGameManager(dbService *services.DatabaseService, analyticsService *services.AnalyticsService) *GameManager {
	gm := &GameManager{
		games:            make(map[string]*models.Game),
		connections:      make(map[*websocket.Conn]*Player),
		waitingQueue:     make([]*Player, 0),
		disconnected:     make(map[string]*DisconnectedInfo),
		dbService:        dbService,
		analyticsService: analyticsService,
		bot:              NewBot(),
	}

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			gm.cleanup()
		}
	}()

	return gm
}

func (gm *GameManager) HandlePlayerJoin(conn *websocket.Conn, data map[string]interface{}) {
	username, ok := data["username"].(string)
	if !ok || len(strings.TrimSpace(username)) < 2 {
		gm.sendError(conn, "Invalid username")
		return
	}
	username = strings.TrimSpace(username)

	gm.mu.Lock()
	defer gm.mu.Unlock()

	// Check for reconnection
	if info, exists := gm.disconnected[username]; exists {
		if time.Since(info.Time).Seconds() <= 30 {
			gm.reconnectPlayer(conn, username, info)
			return
		}
		delete(gm.disconnected, username)
	}

	player := &Player{
		Username: username,
		Conn:     conn,
	}
	gm.connections[conn] = player

	// Try to match with waiting player
	if len(gm.waitingQueue) > 0 {
		opponent := gm.waitingQueue[0]
		gm.waitingQueue = gm.waitingQueue[1:]
		gm.startPvPGame(opponent, player)
		return
	}

	// Add to waiting queue
	gm.waitingQueue = append(gm.waitingQueue, player)
	gm.sendMessage(conn, "waiting_for_opponent", nil)

	// Start bot game after 10 seconds
	go func() {
		time.Sleep(10 * time.Second)
		gm.mu.Lock()
		defer gm.mu.Unlock()
		
		for i, p := range gm.waitingQueue {
			if p == player {
				gm.waitingQueue = append(gm.waitingQueue[:i], gm.waitingQueue[i+1:]...)
				gm.startBotGame(player)
				return
			}
		}
	}()
}

func (gm *GameManager) HandlePlayerMove(conn *websocket.Conn, data map[string]interface{}) {
	gameID, ok := data["gameId"].(string)
	if !ok {
		gm.sendError(conn, "Invalid game ID")
		return
	}

	columnFloat, ok := data["column"].(float64)
	if !ok {
		gm.sendError(conn, "Invalid column")
		return
	}
	column := int(columnFloat)

	gm.mu.Lock()
	defer gm.mu.Unlock()

	player, exists := gm.connections[conn]
	if !exists {
		gm.sendError(conn, "Player not found")
		return
	}

	game, exists := gm.games[gameID]
	if !exists {
		gm.sendError(conn, "Game not found")
		return
	}

	if game.CurrentPlayer != player.PlayerNum {
		gm.sendError(conn, "Not your turn")
		return
	}

	row, gameOver, winner, err := game.MakeMove(column, player.PlayerNum)
	if err != nil {
		gm.sendError(conn, err.Error())
		return
	}

	moveData := map[string]interface{}{
		"column":    column,
		"row":       row,
		"player":    player.PlayerNum,
		"gameState": game,
	}
	gm.broadcastToGame(gameID, "move_made", moveData)

	// Analytics
	if gm.analyticsService != nil {
		gm.analyticsService.TrackEvent("move_made", map[string]interface{}{
			"gameId": gameID,
			"player": player.Username,
			"column": column,
			"row":    row,
		})
	}

	if gameOver {
		gm.endGame(game, winner)
	} else if game.IsBot && game.CurrentPlayer == 2 {
		go func() {
			time.Sleep(1 * time.Second)
			gm.makeBotMove(game)
		}()
	}
}

func (gm *GameManager) HandlePlayerDisconnect(conn *websocket.Conn) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	player, exists := gm.connections[conn]
	if !exists {
		return
	}

	delete(gm.connections, conn)

	// Remove from waiting queue
	for i, p := range gm.waitingQueue {
		if p == player {
			gm.waitingQueue = append(gm.waitingQueue[:i], gm.waitingQueue[i+1:]...)
			return
		}
	}

	// Handle game disconnect
	if player.GameID != "" {
		game, exists := gm.games[player.GameID]
		if exists && game.Status == "playing" {
			gm.disconnected[player.Username] = &DisconnectedInfo{
				GameID:    player.GameID,
				PlayerNum: player.PlayerNum,
				Time:      time.Now(),
			}
			gm.notifyDisconnect(game, player)
		}
	}
}

func (gm *GameManager) startPvPGame(player1, player2 *Player) {
	game := models.NewGame(
		&models.Player{ID: "p1", Username: player1.Username},
		&models.Player{ID: "p2", Username: player2.Username},
	)
	game.Status = "playing"
	gm.games[game.ID] = game

	player1.GameID = game.ID
	player1.PlayerNum = 1
	player2.GameID = game.ID
	player2.PlayerNum = 2

	gm.sendMessage(player1.Conn, "game_started", map[string]interface{}{
		"gameState":  game,
		"yourPlayer": 1,
	})
	gm.sendMessage(player2.Conn, "game_started", map[string]interface{}{
		"gameState":  game,
		"yourPlayer": 2,
	})

	log.Printf("PvP game started: %s vs %s", player1.Username, player2.Username)

	// Analytics
	if gm.analyticsService != nil {
		gm.analyticsService.TrackEvent("game_started", map[string]interface{}{
			"gameId":   game.ID,
			"player1":  player1.Username,
			"player2":  player2.Username,
			"gameType": "pvp",
		})
	}
}

func (gm *GameManager) startBotGame(player *Player) {
	game := models.NewGame(
		&models.Player{ID: "p1", Username: player.Username},
		&models.Player{ID: "bot", Username: "AI Bot", IsBot: true},
	)
	game.Status = "playing"
	game.IsBot = true
	gm.games[game.ID] = game

	player.GameID = game.ID
	player.PlayerNum = 1

	gm.sendMessage(player.Conn, "game_started", map[string]interface{}{
		"gameState":  game,
		"yourPlayer": 1,
	})

	log.Printf("Bot game started for: %s", player.Username)

	// Analytics
	if gm.analyticsService != nil {
		gm.analyticsService.TrackEvent("game_started", map[string]interface{}{
			"gameId":   game.ID,
			"player1":  player.Username,
			"player2":  "AI Bot",
			"gameType": "bot",
		})
	}
}

func (gm *GameManager) reconnectPlayer(conn *websocket.Conn, username string, info *DisconnectedInfo) {
	game, exists := gm.games[info.GameID]
	if !exists || game.Status != "playing" {
		delete(gm.disconnected, username)
		gm.sendError(conn, "Game no longer available")
		return
	}

	player := &Player{
		Username:  username,
		Conn:      conn,
		GameID:    info.GameID,
		PlayerNum: info.PlayerNum,
	}
	gm.connections[conn] = player
	delete(gm.disconnected, username)

	gm.sendMessage(conn, "game_rejoined", map[string]interface{}{
		"gameState":  game,
		"yourPlayer": info.PlayerNum,
	})

	gm.notifyReconnect(game, player)
	log.Printf("Player %s reconnected to game %s", username, info.GameID)
}

func (gm *GameManager) makeBotMove(game *models.Game) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if game.Status != "playing" || game.CurrentPlayer != 2 {
		return
	}

	column := gm.bot.GetBestMove(game)
	if column < 0 || column > 6 {
		log.Printf("Bot selected invalid column: %d", column)
		return
	}

	// Check if column is full
	if game.Board[0][column] != 0 {
		log.Printf("Bot selected full column: %d, finding alternative", column)
		// Find first available column
		for col := 0; col < 7; col++ {
			if game.Board[0][col] == 0 {
				column = col
				break
			}
		}
	}

	row, gameOver, winner, err := game.MakeMove(column, 2)
	if err != nil {
		log.Printf("Bot move error: %v", err)
		return
	}

	moveData := map[string]interface{}{
		"column":    column,
		"row":       row,
		"player":    2,
		"gameState": game,
	}
	gm.broadcastToGame(game.ID, "move_made", moveData)

	// Analytics
	if gm.analyticsService != nil {
		gm.analyticsService.TrackEvent("bot_move", map[string]interface{}{
			"gameId": game.ID,
			"column": column,
			"row":    row,
		})
	}

	if gameOver {
		gm.endGame(game, winner)
	}
}

func (gm *GameManager) endGame(game *models.Game, winner *int) {
	game.Status = "finished"
	game.Winner = winner

	endData := map[string]interface{}{
		"winner":    winner,
		"gameState": game,
	}
	gm.broadcastToGame(game.ID, "game_ended", endData)

	// Analytics
	if gm.analyticsService != nil {
		gm.analyticsService.TrackEvent("game_ended", map[string]interface{}{
			"gameId":   game.ID,
			"winner":   winner,
			"duration": game.GetDuration(),
			"moves":    len(game.Moves),
			"gameType": map[bool]string{true: "bot", false: "pvp"}[game.IsBot],
		})
	}

	// Save to database
	go func() {
		if gm.dbService != nil {
			gameData := services.GameData{
				ID:       game.ID,
				Player1:  game.Player1.Username,
				Player2:  game.Player2.Username,
				Winner:   winner,
				Duration: game.GetDuration(),
				Moves:    len(game.Moves),
				IsBot:    game.IsBot,
			}
			gm.dbService.SaveGame(gameData)

			if winner != nil {
				winnerName := game.Player1.Username
				if *winner == 2 {
					winnerName = game.Player2.Username
				}
				gm.dbService.UpdatePlayerStats(winnerName, true)
			}
		}
	}()

	// Cleanup after 30 seconds
	go func() {
		time.Sleep(30 * time.Second)
		gm.mu.Lock()
		delete(gm.games, game.ID)
		gm.mu.Unlock()
	}()
}

func (gm *GameManager) cleanup() {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	now := time.Now()
	for username, info := range gm.disconnected {
		if now.Sub(info.Time).Seconds() > 30 {
			game, exists := gm.games[info.GameID]
			if exists && game.Status == "playing" {
				winner := 2
				if info.PlayerNum == 2 {
					winner = 1
				}
				gm.endGame(game, &winner)
			}
			delete(gm.disconnected, username)
		}
	}
}

func (gm *GameManager) broadcastToGame(gameID string, msgType string, data interface{}) {
	for conn, player := range gm.connections {
		if player.GameID == gameID {
			gm.sendMessage(conn, msgType, data)
		}
	}
}

func (gm *GameManager) notifyDisconnect(game *models.Game, player *Player) {
	for conn, p := range gm.connections {
		if p.GameID == game.ID && p != player {
			gm.sendMessage(conn, "player_disconnected", map[string]interface{}{
				"player":        player.Username,
				"reconnectTime": 30,
			})
		}
	}
}

func (gm *GameManager) notifyReconnect(game *models.Game, player *Player) {
	for conn, p := range gm.connections {
		if p.GameID == game.ID && p != player {
			gm.sendMessage(conn, "player_reconnected", map[string]interface{}{
				"player": player.Username,
			})
		}
	}
}

func (gm *GameManager) sendMessage(conn *websocket.Conn, msgType string, data interface{}) {
	message := map[string]interface{}{
		"type": msgType,
		"data": data,
	}
	if err := conn.WriteJSON(message); err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

func (gm *GameManager) sendError(conn *websocket.Conn, message string) {
	gm.sendMessage(conn, "error", map[string]string{"message": message})
}