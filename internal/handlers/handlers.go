package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"emitrr-4-in-a-row/internal/game"
	"emitrr-4-in-a-row/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	gameManager *game.GameManager
	dbService   *services.DatabaseService
	upgrader    websocket.Upgrader
}

func NewHandler(gameManager *game.GameManager, dbService *services.DatabaseService) *Handler {
	return &Handler{
		gameManager: gameManager,
		dbService:   dbService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins
			},
			HandshakeTimeout: 10 * time.Second,
		},
	}
}

func (h *Handler) SetupRoutes(router *gin.Engine) {
	// API routes
	api := router.Group("/api")
	{
		api.GET("/leaderboard", h.getLeaderboard)
		api.GET("/analytics", h.getAnalytics)
	}

	// WebSocket endpoint
	router.GET("/ws", h.handleWebSocket)

	// Serve static files (React build)
	router.Static("/static", "./frontend/build/static")
	router.StaticFile("/favicon.ico", "./frontend/build/favicon.ico")
	router.StaticFile("/manifest.json", "./frontend/build/manifest.json")

	// Serve React app for all other routes
	router.NoRoute(func(c *gin.Context) {
		// Check if it's an API route
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(404, gin.H{"error": "Not found"})
			return
		}
		if len(c.Request.URL.Path) >= 3 && c.Request.URL.Path[:3] == "/ws" {
			c.JSON(404, gin.H{"error": "Not found"})
			return
		}

		// Serve index.html for React routing
		c.File("./frontend/build/index.html")
	})
}

func (h *Handler) getLeaderboard(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	leaderboard, err := h.dbService.GetLeaderboard(limit)
	if err != nil {
		// Return empty leaderboard if DB unavailable
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	c.JSON(http.StatusOK, leaderboard)
}

func (h *Handler) getAnalytics(c *gin.Context) {
	analytics, err := h.dbService.GetAnalytics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch analytics"})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

func (h *Handler) handleWebSocket(c *gin.Context) {
	log.Printf("WebSocket upgrade attempt from: %s", c.Request.RemoteAddr)
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("Player connected successfully: %s", conn.RemoteAddr())
	
	// Set read/write timeouts
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	// Handle messages
	for {
		var message map[string]interface{}
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		messageType, ok := message["type"].(string)
		if !ok {
			continue
		}

		data, ok := message["data"].(map[string]interface{})
		if !ok {
			data = make(map[string]interface{})
		}

		switch messageType {
		case "join_game":
			h.gameManager.HandlePlayerJoin(conn, data)
		case "make_move":
			h.gameManager.HandlePlayerMove(conn, data)
		case "rejoin_game":
			gameID, _ := data["gameId"].(string)
			username, _ := data["username"].(string)
			if gameID != "" && username != "" {
				h.gameManager.HandlePlayerRejoin(conn, gameID, username)
			}
		case "ping":
			// Keepalive ping - just reset read deadline
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		default:
			log.Printf("Unknown message type: %s", messageType)
		}
	}

	// Handle disconnect
	h.gameManager.HandlePlayerDisconnect(conn)
	log.Printf("Player disconnected: %s", conn.RemoteAddr())
}