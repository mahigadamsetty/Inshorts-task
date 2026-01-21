package main

import (
	"fmt"
	"log"

	"github.com/mahigadamsetty/Inshorts-task/internal/config"
	"github.com/mahigadamsetty/Inshorts-task/internal/db"
	"github.com/mahigadamsetty/Inshorts-task/internal/models"
	"github.com/mahigadamsetty/Inshorts-task/internal/services"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	if err := db.Init(cfg.DatabaseURL); err != nil {
		log.Fatalf("could not initialize database: %v", err)
	}

	fmt.Println("Database initialized.")

	// Fetch all article IDs to simulate events for
	var articles []models.Article
	database := db.GetDB()
	if err := database.Select("id, latitude, longitude").Find(&articles).Error; err != nil {
		log.Fatalf("could not fetch articles: %v", err)
	}

	if len(articles) == 0 {
		log.Fatal("No articles found in the database. Please import data first.")
	}

	// Number of events to simulate
	eventCount := 1000

	fmt.Printf("Simulating %d user events...\n", eventCount)

	// Simulate events
	if err := services.SimulateUserEvents(articles, eventCount); err != nil {
		log.Fatalf("could not simulate user events: %v", err)
	}

	fmt.Println("Successfully simulated user events.")
}
