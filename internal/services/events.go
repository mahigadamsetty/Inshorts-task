package services

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mahigadamsetty/Inshorts-task/internal/db"
	"github.com/mahigadamsetty/Inshorts-task/internal/models"
)

// SimulateUserEvents creates a specified number of random user events (views/clicks)
// for a given list of articles.
func SimulateUserEvents(articles []models.Article, count int) error {
	database := db.GetDB()
	if database == nil {
		return fmt.Errorf("database not initialized")
	}

	for i := 0; i < count; i++ {
		// Pick a random article
		article := articles[rand.Intn(len(articles))]

		// Simulate a user location near the article's location
		userLat := article.Latitude + (rand.Float64()-0.5)*0.5 // within ~55km
		userLon := article.Longitude + (rand.Float64()-0.5)*0.5

		// Decide event type (80% view, 20% click)
		eventType := models.EventTypeView
		if rand.Float64() < 0.2 {
			eventType = models.EventTypeClick
		}

		event := models.Event{
			ArticleID: article.ID,
			EventType: eventType,
			Latitude:  userLat,
			Longitude: userLon,
			Timestamp: time.Now(),
		}

		if err := database.Create(&event).Error; err != nil {
			// Log or handle individual event creation errors if necessary,
			// but continue simulating other events.
		}
	}

	return nil
}
