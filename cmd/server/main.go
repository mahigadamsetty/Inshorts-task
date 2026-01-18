package main

import (
	"log"

	"github.com/mahigadamsetty/Inshorts-task/internal/config"
	"github.com/mahigadamsetty/Inshorts-task/internal/db"
	"github.com/mahigadamsetty/Inshorts-task/internal/router"
	"github.com/mahigadamsetty/Inshorts-task/internal/services"
)

func main() {
	// Load configuration
	cfg := config.Load()
	
	// Initialize database
	if err := db.Init(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	
	// Initialize trending cache
	services.InitTrendingCache(cfg.TrendingCacheTTL)
	
	// Setup router
	r := router.SetupRouter(cfg)
	
	// Start server
	addr := ":" + cfg.Port
	log.Printf("Starting server on %s", addr)
	log.Printf("OpenAI API Key configured: %v", cfg.OpenAIAPIKey != "")
	log.Printf("LLM Model: %s", cfg.LLMModel)
	
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
