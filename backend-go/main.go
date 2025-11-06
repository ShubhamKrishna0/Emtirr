package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"four-in-a-row/internal/config"
	"four-in-a-row/internal/game"
	"four-in-a-row/internal/handlers"
	"four-in-a-row/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	// Initialize services
	dbService := services.NewDatabaseService(cfg)
	analyticsService := services.NewAnalyticsService(cfg)
	gameManager := game.NewGameManager(dbService, analyticsService)

	// Initialize services
	if err := analyticsService.Initialize(); err != nil {
		log.Printf("Analytics initialization failed: %v", err)
	}

	if err := dbService.Initialize(); err != nil {
		log.Printf("Database unavailable, continuing without persistence: %v", err)
	}

	// Setup router
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	// Setup handlers
	h := handlers.NewHandler(gameManager, dbService)
	h.SetupRoutes(router)

	// Start analytics consumer
	go func() {
		if err := analyticsService.StartConsumer(dbService); err != nil {
			log.Printf("Analytics consumer failed: %v", err)
		}
	}()

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Server running on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	analyticsService.Close()
	dbService.Close()
	log.Println("Server exited")
}