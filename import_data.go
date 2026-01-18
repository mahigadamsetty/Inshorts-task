package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/mahigadamsetty/Inshorts-task/internal/config"
	"github.com/mahigadamsetty/Inshorts-task/internal/db"
	"github.com/mahigadamsetty/Inshorts-task/internal/models"
	"github.com/mahigadamsetty/Inshorts-task/internal/services"
)

type JSONArticle struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	URL             string    `json:"url"`
	PublicationDate string    `json:"publication_date"`
	SourceName      string    `json:"source_name"`
	Category        []string  `json:"category"`
	RelevanceScore  float64   `json:"relevance_score"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run import_data.go <path_to_json_file>")
	}

	filename := os.Args[1]

	// Load configuration
	cfg := config.Load()

	// Initialize database
	if err := db.Init(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Read JSON file
	log.Printf("Reading file: %s", filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Parse JSON
	var jsonArticles []JSONArticle
	if err := json.Unmarshal(data, &jsonArticles); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	log.Printf("Found %d articles to import", len(jsonArticles))

	// Convert to GORM models
	articles := make([]models.Article, len(jsonArticles))
	for i, ja := range jsonArticles {
		// Parse publication date
		pubDate, err := time.Parse("2006-01-02T15:04:05", ja.PublicationDate)
		if err != nil {
			// Try alternative formats
			pubDate, err = time.Parse(time.RFC3339, ja.PublicationDate)
			if err != nil {
				log.Printf("Warning: Failed to parse date for article %s: %v", ja.ID, err)
				pubDate = time.Now()
			}
		}

		articles[i] = models.Article{
			ID:              ja.ID,
			Title:           ja.Title,
			Description:     ja.Description,
			URL:             ja.URL,
			PublicationDate: pubDate,
			SourceName:      ja.SourceName,
			Category:        models.StringArray(ja.Category),
			RelevanceScore:  ja.RelevanceScore,
			Latitude:        ja.Latitude,
			Longitude:       ja.Longitude,
		}
	}

	// Import in batches
	batchSize := 100
	database := db.GetDB()
	
	for i := 0; i < len(articles); i += batchSize {
		end := i + batchSize
		if end > len(articles) {
			end = len(articles)
		}

		batch := articles[i:end]
		if err := database.Create(&batch).Error; err != nil {
			log.Printf("Warning: Failed to import batch %d-%d: %v", i, end, err)
		} else {
			log.Printf("Imported articles %d-%d", i, end)
		}
	}

	log.Println("Import complete!")

	// Simulate some user events for trending
	log.Println("Simulating user events for trending...")
	articleIDs := make([]string, len(articles))
	for i, article := range articles {
		articleIDs[i] = article.ID
	}

	// Generate 1000 random events
	if err := services.SimulateUserEvents(articleIDs, 1000); err != nil {
		log.Printf("Warning: Failed to simulate events: %v", err)
	} else {
		log.Println("Simulated 1000 user events")
	}

	// Print summary
	var count int64
	database.Model(&models.Article{}).Count(&count)
	fmt.Printf("\nDatabase now contains %d articles\n", count)
	
	var eventCount int64
	database.Model(&models.Event{}).Count(&eventCount)
	fmt.Printf("Database now contains %d events\n", eventCount)
}
