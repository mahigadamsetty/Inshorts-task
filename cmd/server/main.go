package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mahigadamsetty/Inshorts-task/internal/config"
	"github.com/mahigadamsetty/Inshorts-task/internal/db"
	"github.com/mahigadamsetty/Inshorts-task/internal/handlers"
	"github.com/mahigadamsetty/Inshorts-task/internal/llm"
	"github.com/mahigadamsetty/Inshorts-task/internal/router"
	"github.com/mahigadamsetty/Inshorts-task/internal/services"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	if err := db.Initialize(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database initialized successfully")

	// Initialize LLM client
	llmClient := llm.NewClient(cfg.OpenAIAPIKey, cfg.LLMModel)
	if cfg.OpenAIAPIKey == "" {
		log.Println("Warning: No OpenAI API key provided. Using fallback heuristics.")
	} else {
		log.Println("LLM client initialized with OpenAI")
	}

	// Initialize trending service
	trendingService := services.NewTrendingService(cfg.TrendingCacheTTL, cfg.LocationClusterDegrees)
	
	// Start event simulation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	trendingService.StartEventSimulation(ctx)
	log.Println("Event simulation started")

	// Initialize handlers
	newsHandler := handlers.NewNewsHandler(llmClient, trendingService)

	// Setup router
	r := router.Setup(newsHandler)

	// Start server
	addr := ":" + cfg.Port
	log.Printf("Starting server on %s", addr)
	log.Printf("API endpoints available at http://localhost%s/api/v1/news/", addr)

	// Graceful shutdown
	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	cancel()
	log.Println("Server stopped")
}
