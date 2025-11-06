package game

import (
	"log"
	"strings"
	"sync"
	"time"

	"four-in-a-row/internal/models"
	"four-in-a-row/internal/services"

	"github.com/gorilla/websocket"
)

type GameManager struct {
	games               map[string]*models.Game
	waitingPlayers      map[string]*PlayerConnection
	playerSockets       map[string]*PlayerConnection
	disconnectedPlayers map[string]*DisconnectedPlayer
	dbService           *services.DatabaseService
	analyticsService    *services.AnalyticsService
	bot                 *Bot
	mu                  sync.RWMutex
}

type PlayerConnection struct {
	Player *models.Player
	Conn   *websocket.Conn
}

type DisconnectedPlayer struct {
	GameID         string
	DisconnectedAt time.Time
}

type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func NewGameManager(dbService *services.DatabaseService, analyticsService *services.AnalyticsService) *GameManager {
	gm := &GameManager{
		games:               make(map[string]*models.Game),
		waitingPlayers:      make(map[string]*PlayerConnection),
		playerSockets:       make(map[string]*PlayerConnection),
		disconnectedPlayers: make(map[string]*DisconnectedPlayer),
		dbService:           dbService,
		analyticsService:    analyticsService,
		bot:                 NewBot(),
	}

	// Cleanup disconnected players every 30 seconds
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			gm.cleanupDisconnectedPlayers()
		}
	}()

	return gm
}

func (gm *GameManager) HandlePlayerJoin(conn *websocket.Conn, data map[string]interface{}) {
	log.Printf("HandlePlayerJoin called with data: %+v", data)
	username, ok := data["username"].(string)
	if !ok {
		log.Printf("Username not found or not string: %+v", data["username"])
		gm.sendError(conn, "Username is required")
		return
	}
	username = strings.TrimSpace(username)
	if len(username) < 2 || len(username) > 20 {
		log.Printf("Invalid username length: %s (len=%d)", username, len(username))
		gm.sendError(conn, "Username must be 2-20 characters")
		return
	}
	log.Printf("Player joining: %s", username)

	player := &models.Player{
		ID:       generatePlayerID(),
		Username: username,
		IsBot:    false,
	}

	playerConn := &PlayerConnection{
		Player: player,
		Conn:   conn,
	}

	gm.mu.Lock()
	gm.playerSockets[player.ID] = playerConn

	// Check for reconnectable game
	if reconnectGame := gm.findReconnectableGame(username); reconnectGame != nil {
		gm.mu.Unlock()
		gm.HandlePlayerRejoin(conn, reconnectGame.ID, username)
		return
	}

	// Try to match with waiting player
	waitingPlayer := gm.findWaitingPlayer(username)
	if waitingPlayer != nil {
		delete(gm.waitingPlayers, waitingPlayer.Player.ID)
		gm.mu.Unlock()
		gm.createGame(waitingPlayer, playerConn)
	} else {
		gm.waitingPlayers[player.ID] = playerConn
		gm.mu.Unlock()
		log.Printf("Sending waiting_for_opponent to player %s", username)
		gm.sendMessage(conn, "waiting_for_opponent", nil)

		// Start bot game after 10 seconds
		go func() {
			time.Sleep(10 * time.Second)
			gm.mu.Lock()
			if _, exists := gm.waitingPlayers[player.ID]; exists {
				delete(gm.waitingPlayers, player.ID)
				gm.mu.Unlock()
				gm.createBotGame(playerConn)
			} else {
				gm.mu.Unlock()
			}
		}()
	}
}

func (gm *GameManager) findWaitingPlayer(currentUsername string) *PlayerConnection {
	for id, playerConn := range gm.waitingPlayers {
		if playerConn.Player.Username != currentUsername {
			delete(gm.waitingPlayers, id)
			return playerConn
		}
	}
	return nil
}

func (gm *GameManager) createGame(player1, player2 *PlayerConnection) {
	game := models.NewGame(player1.Player, player2.Player)
	game.Status = "playing"

	gm.mu.Lock()
	gm.games[game.ID] = game
	gm.mu.Unlock()

	gameState := map[string]interface{}{
		"gameId":    game.ID,
		"gameState": game,
		"yourPlayer": 1,
	}

	gm.sendMessage(player1.Conn, "game_started", gameState)
	gameState["yourPlayer"] = 2
	gm.sendMessage(player2.Conn, "game_started", gameState)

	gm.sendMessage(player1.Conn, "your_turn", map[string]int{"player": 1})
	gm.sendMessage(player2.Conn, "your_turn", map[string]int{"player": 2})

	gm.analyticsService.TrackEvent("game_started", map[string]interface{}{
		"gameId":   game.ID,
		"player1":  player1.Player.Username,
		"player2":  player2.Player.Username,
		"gameType": "pvp",
	})
}

func (gm *GameManager) createBotGame(playerConn *PlayerConnection) {
	botPlayer := &models.Player{
		ID:       gm.bot.ID,
		Username: gm.bot.Username,
		IsBot:    true,
	}

	game := models.NewGame(playerConn.Player, botPlayer)
	game.Status = "playing"
	game.IsBot = true

	gm.mu.Lock()
	gm.games[game.ID] = game
	gm.mu.Unlock()

	gameState := map[string]interface{}{
		"gameId":     game.ID,
		"gameState":  game,
		"yourPlayer": 1,
	}

	log.Printf("Sending game_started to player %s", playerConn.Player.Username)
	gm.sendMessage(playerConn.Conn, "game_started", gameState)

	gm.analyticsService.TrackEvent("game_started", map[string]interface{}{
		"gameId":   game.ID,
		"player1":  playerConn.Player.Username,
		"player2":  "AI Bot",
		"gameType": "bot",
	})
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

	playerID := gm.getPlayerIDByConn(conn)
	if playerID == "" {
		gm.sendError(conn, "Player not found")
		return
	}

	gm.mu.RLock()
	game, exists := gm.games[gameID]
	gm.mu.RUnlock()

	if !exists {
		gm.sendError(conn, "Game not found")
		return
	}

	row, gameOver, _, err := game.MakeMove(column, playerID)
	if err != nil {
		gm.sendError(conn, err.Error())
		return
	}

	moveData := map[string]interface{}{
		"column":    column,
		"row":       row,
		"player":    game.GetPlayerNumber(playerID),
		"gameState": game,
	}

	gm.broadcastToGame(gameID, "move_made", moveData)

	gm.analyticsService.TrackEvent("move_made", map[string]interface{}{
		"gameId": gameID,
		"player": playerID,
		"column": column,
		"row":    row,
	})

	if gameOver {
		gm.handleGameEnd(game)
	} else if game.IsBot && game.CurrentPlayer == 2 {
		go func() {
			time.Sleep(1 * time.Second)
			gm.makeBotMove(game)
		}()
	}
}

func (gm *GameManager) makeBotMove(game *models.Game) {
	if game.Status != "playing" || game.CurrentPlayer != 2 {
		return
	}

	var column int
	if immediateMove := gm.bot.GetImmediateMove(game); immediateMove != nil {
		column = *immediateMove
	} else {
		column = gm.bot.GetBestMove(game)
	}

	row, gameOver, _, err := game.MakeMove(column, gm.bot.ID)
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

	gm.analyticsService.TrackEvent("bot_move", map[string]interface{}{
		"gameId": game.ID,
		"column": column,
		"row":    row,
	})

	if gameOver {
		gm.handleGameEnd(game)
	}
}

func (gm *GameManager) handleGameEnd(game *models.Game) {
	duration := game.GetDuration()

	// Save to database
	go func() {
		gameData := services.GameData{
			ID:        game.ID,
			Player1:   game.Player1.Username,
			Player2:   game.Player2.Username,
			Winner:    game.Winner,
			Duration:  duration,
			Moves:     len(game.Moves),
			IsBot:     game.IsBot,
			CreatedAt: game.CreatedAt,
		}

		if err := gm.dbService.SaveGame(gameData); err != nil {
			log.Printf("Failed to save game: %v", err)
		}

		if game.Winner != nil {
			winnerUsername := game.Player1.Username
			if *game.Winner == 2 {
				winnerUsername = game.Player2.Username
			}
			gm.dbService.UpdatePlayerStats(winnerUsername, true)

			if !game.IsBot {
				loserUsername := game.Player2.Username
				if *game.Winner == 2 {
					loserUsername = game.Player1.Username
				}
				gm.dbService.UpdatePlayerStats(loserUsername, false)
			}
		}
	}()

	endData := map[string]interface{}{
		"winner":    game.Winner,
		"gameState": game,
		"duration":  duration,
	}

	gm.broadcastToGame(game.ID, "game_ended", endData)

	gm.analyticsService.TrackEvent("game_ended", map[string]interface{}{
		"gameId":   game.ID,
		"winner":   game.Winner,
		"duration": duration,
		"moves":    len(game.Moves),
		"gameType": map[bool]string{true: "bot", false: "pvp"}[game.IsBot],
	})

	// Cleanup after 30 seconds
	go func() {
		time.Sleep(30 * time.Second)
		gm.mu.Lock()
		delete(gm.games, game.ID)
		gm.mu.Unlock()
	}()
}

func (gm *GameManager) HandlePlayerRejoin(conn *websocket.Conn, gameID, username string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	game, exists := gm.games[gameID]
	if !exists {
		gm.sendError(conn, "No reconnectable game found")
		return
	}

	playerID := generatePlayerID()
	if game.Player1.Username == username {
		game.Player1.ID = playerID
		gm.playerSockets[playerID] = &PlayerConnection{Player: game.Player1, Conn: conn}
	} else if game.Player2.Username == username {
		game.Player2.ID = playerID
		gm.playerSockets[playerID] = &PlayerConnection{Player: game.Player2, Conn: conn}
	}

	delete(gm.disconnectedPlayers, username)

	rejoinData := map[string]interface{}{
		"gameId":     game.ID,
		"gameState":  game,
		"yourPlayer": game.GetPlayerNumber(playerID),
	}

	gm.sendMessage(conn, "game_rejoined", rejoinData)

	gm.analyticsService.TrackEvent("player_rejoined", map[string]interface{}{
		"gameId": game.ID,
		"player": username,
	})
}

func (gm *GameManager) HandlePlayerDisconnect(conn *websocket.Conn) {
	playerID := gm.getPlayerIDByConn(conn)
	if playerID == "" {
		return
	}

	gm.mu.Lock()
	defer gm.mu.Unlock()

	playerConn, exists := gm.playerSockets[playerID]
	if !exists {
		return
	}

	delete(gm.waitingPlayers, playerID)
	delete(gm.playerSockets, playerID)

	game := gm.findPlayerGame(playerID)
	if game != nil && game.Status == "playing" {
		gm.disconnectedPlayers[playerConn.Player.Username] = &DisconnectedPlayer{
			GameID:         game.ID,
			DisconnectedAt: time.Now(),
		}

		gm.broadcastToGameExcept(game.ID, playerID, "player_disconnected", map[string]interface{}{
			"player":        playerConn.Player.Username,
			"reconnectTime": 30,
		})

		gm.analyticsService.TrackEvent("player_disconnected", map[string]interface{}{
			"gameId": game.ID,
			"player": playerConn.Player.Username,
		})
	}
}

func (gm *GameManager) findPlayerGame(playerID string) *models.Game {
	for _, game := range gm.games {
		if (game.Player1.ID == playerID) || (game.Player2 != nil && game.Player2.ID == playerID) {
			return game
		}
	}
	return nil
}

func (gm *GameManager) findReconnectableGame(username string) *models.Game {
	disconnectedInfo, exists := gm.disconnectedPlayers[username]
	if !exists {
		return nil
	}

	game, exists := gm.games[disconnectedInfo.GameID]
	if !exists || game.Status != "playing" {
		return nil
	}

	return game
}

func (gm *GameManager) cleanupDisconnectedPlayers() {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	now := time.Now()
	for username, info := range gm.disconnectedPlayers {
		if now.Sub(info.DisconnectedAt).Seconds() > 30 {
			game, exists := gm.games[info.GameID]
			if exists && game.Status == "playing" {
				disconnectedPlayer := 1
				if game.Player1.Username == username {
					disconnectedPlayer = 1
				} else {
					disconnectedPlayer = 2
				}

				winner := 2
				if disconnectedPlayer == 2 {
					winner = 1
				}
				game.Winner = &winner
				game.Status = "finished"

				gm.handleGameEnd(game)
			}
			delete(gm.disconnectedPlayers, username)
		}
	}
}

func (gm *GameManager) getPlayerIDByConn(conn *websocket.Conn) string {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	for id, playerConn := range gm.playerSockets {
		if playerConn.Conn == conn {
			return id
		}
	}
	return ""
}

func (gm *GameManager) broadcastToGame(gameID, messageType string, data interface{}) {
	gm.mu.RLock()
	game, exists := gm.games[gameID]
	gm.mu.RUnlock()

	if !exists {
		return
	}

	if playerConn, exists := gm.playerSockets[game.Player1.ID]; exists {
		gm.sendMessage(playerConn.Conn, messageType, data)
	}

	if game.Player2 != nil && !game.Player2.IsBot {
		if playerConn, exists := gm.playerSockets[game.Player2.ID]; exists {
			gm.sendMessage(playerConn.Conn, messageType, data)
		}
	}
}

func (gm *GameManager) broadcastToGameExcept(gameID, exceptPlayerID, messageType string, data interface{}) {
	gm.mu.RLock()
	game, exists := gm.games[gameID]
	gm.mu.RUnlock()

	if !exists {
		return
	}

	if game.Player1.ID != exceptPlayerID {
		if playerConn, exists := gm.playerSockets[game.Player1.ID]; exists {
			gm.sendMessage(playerConn.Conn, messageType, data)
		}
	}

	if game.Player2 != nil && game.Player2.ID != exceptPlayerID && !game.Player2.IsBot {
		if playerConn, exists := gm.playerSockets[game.Player2.ID]; exists {
			gm.sendMessage(playerConn.Conn, messageType, data)
		}
	}
}

func (gm *GameManager) sendMessage(conn *websocket.Conn, messageType string, data interface{}) {
	message := map[string]interface{}{
		"type": messageType,
		"data": data,
	}

	log.Printf("Sending WebSocket message: type=%s, data=%+v", messageType, data)
	if err := conn.WriteJSON(message); err != nil {
		log.Printf("Failed to send message %s: %v", messageType, err)
	} else {
		log.Printf("Successfully sent message: %s", messageType)
	}
}

func (gm *GameManager) sendError(conn *websocket.Conn, message string) {
	gm.sendMessage(conn, "error", map[string]string{"message": message})
}

func generatePlayerID() string {
	return time.Now().Format("20060102150405") + "-" + string(rune(time.Now().Nanosecond()%1000))
}