package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mahigadamsetty/Inshorts-task/internal/config"
	"github.com/mahigadamsetty/Inshorts-task/internal/db"
	"github.com/mahigadamsetty/Inshorts-task/internal/models"
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
		fmt.Println("Usage: go run import_data.go <json_file>")
		fmt.Println("Example: go run import_data.go \"news_data (1).json\"")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Load configuration
	cfg := config.Load()

	// Initialize database
	if err := db.Initialize(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database initialized successfully")

	// Read JSON file
	log.Printf("Reading data from %s...", filename)
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Parse JSON
	var jsonArticles []JSONArticle
	if err := json.Unmarshal(data, &jsonArticles); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	log.Printf("Found %d articles in JSON file", len(jsonArticles))

	// Convert and insert articles
	successCount := 0
	errorCount := 0

	for i, ja := range jsonArticles {
		// Parse publication date
		pubDate, err := time.Parse("2006-01-02T15:04:05", ja.PublicationDate)
		if err != nil {
			// Try alternative format
			pubDate, err = time.Parse(time.RFC3339, ja.PublicationDate)
			if err != nil {
				log.Printf("Warning: Failed to parse date for article %s: %v", ja.ID, err)
				pubDate = time.Now()
			}
		}

		// Convert category array to comma-separated string
		categoryStr := strings.Join(ja.Category, ",")

		article := models.Article{
			ID:              ja.ID,
			Title:           ja.Title,
			Description:     ja.Description,
			URL:             ja.URL,
			PublicationDate: pubDate,
			SourceName:      ja.SourceName,
			Category:        categoryStr,
			CategoryArray:   ja.Category,
			RelevanceScore:  ja.RelevanceScore,
			Latitude:        ja.Latitude,
			Longitude:       ja.Longitude,
		}

		// Insert or update article
		result := db.GetDB().Save(&article)
		if result.Error != nil {
			log.Printf("Error inserting article %s: %v", ja.ID, result.Error)
			errorCount++
		} else {
			successCount++
		}

		// Progress indicator
		if (i+1)%1000 == 0 {
			log.Printf("Processed %d/%d articles...", i+1, len(jsonArticles))
		}
	}

	log.Printf("\nImport completed!")
	log.Printf("Successfully imported: %d articles", successCount)
	if errorCount > 0 {
		log.Printf("Failed to import: %d articles", errorCount)
	}
}
