package services

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/mahigadamsetty/Inshorts-task/internal/db"
	"github.com/mahigadamsetty/Inshorts-task/internal/models"
	"github.com/mahigadamsetty/Inshorts-task/internal/utils"
)

// TrendingCache stores trending results by location cluster
type TrendingCache struct {
	cache  map[string]*CacheEntry
	mu     sync.RWMutex
	ttl    time.Duration
	ticker *time.Ticker
}

type CacheEntry struct {
	Articles  []models.Article
	Timestamp time.Time
}

var trendingCache *TrendingCache

// InitTrendingCache initializes the trending cache
func InitTrendingCache(ttl int) {
	trendingCache = &TrendingCache{
		cache:  make(map[string]*CacheEntry),
		ttl:    time.Duration(ttl) * time.Second,
		ticker: time.NewTicker(time.Duration(ttl) * time.Second),
	}
	
	// Start cleanup goroutine
	go trendingCache.cleanup()
}

// cleanup periodically removes expired cache entries
func (tc *TrendingCache) cleanup() {
	for range tc.ticker.C {
		tc.mu.Lock()
		now := time.Now()
		for key, entry := range tc.cache {
			if now.Sub(entry.Timestamp) > tc.ttl {
				delete(tc.cache, key)
			}
		}
		tc.mu.Unlock()
	}
}

// Get retrieves cached trending articles for a location cluster
func (tc *TrendingCache) Get(key string) ([]models.Article, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	
	entry, exists := tc.cache[key]
	if !exists {
		return nil, false
	}
	
	// Check if cache entry is still valid
	if time.Since(entry.Timestamp) > tc.ttl {
		return nil, false
	}
	
	return entry.Articles, true
}

// Set stores trending articles for a location cluster
func (tc *TrendingCache) Set(key string, articles []models.Article) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	
	tc.cache[key] = &CacheEntry{
		Articles:  articles,
		Timestamp: time.Now(),
	}
}

// SimulateUserEvents generates random user events for articles
func SimulateUserEvents(articleIDs []string, count int) error {
	database := db.GetDB()
	rand.Seed(time.Now().UnixNano())
	
	events := make([]models.Event, count)
	
	for i := 0; i < count; i++ {
		// Random article
		articleID := articleIDs[rand.Intn(len(articleIDs))]
		
		// Random event type (70% views, 30% clicks)
		eventType := models.EventTypeView
		if rand.Float64() < 0.3 {
			eventType = models.EventTypeClick
		}
		
		// Random location (simulate various locations)
		lat := 20.0 + rand.Float64()*20.0  // Range: 20-40
		lon := 70.0 + rand.Float64()*20.0  // Range: 70-90
		
		// Random timestamp within last 24 hours
		hoursAgo := rand.Float64() * 24
		timestamp := time.Now().Add(-time.Duration(hoursAgo) * time.Hour)
		
		events[i] = models.Event{
			ArticleID: articleID,
			EventType: eventType,
			Latitude:  lat,
			Longitude: lon,
			Timestamp: timestamp,
		}
	}
	
	// Batch insert
	return database.Create(&events).Error
}

// ArticleScore represents an article with its trending score
type ArticleScore struct {
	ArticleID string
	Score     float64
}

// GetTrendingArticles calculates and returns trending articles for a location
func GetTrendingArticles(lat, lon float64, limit int, clusterDegrees float64) ([]models.Article, error) {
	// Check cache first
	clusterKey := utils.GetLocationClusterKey(lat, lon, clusterDegrees)
	if cached, found := trendingCache.Get(clusterKey); found {
		if len(cached) <= limit {
			return cached, nil
		}
		return cached[:limit], nil
	}
	
	database := db.GetDB()
	
	// Get recent events (last 24 hours)
	cutoffTime := time.Now().Add(-24 * time.Hour)
	var events []models.Event
	
	if err := database.Where("timestamp > ?", cutoffTime).Find(&events).Error; err != nil {
		return nil, err
	}
	
	// Calculate trending scores
	articleScores := make(map[string]float64)
	
	for _, event := range events {
		// Calculate distance from user location
		distance := utils.HaversineDistance(lat, lon, event.Latitude, event.Longitude)
		
		// Calculate recency factor (decay over time)
		hoursAgo := time.Since(event.Timestamp).Hours()
		recencyFactor := math.Exp(-hoursAgo / 12.0) // Exponential decay, half-life of 12 hours
		
		// Calculate interaction weight (clicks worth more than views)
		interactionWeight := 1.0
		if event.EventType == models.EventTypeClick {
			interactionWeight = 2.0
		}
		
		// Calculate geographical relevance (closer locations score higher)
		// Use inverse distance with smoothing
		geoRelevance := 1.0 / (1.0 + distance/100.0)
		
		// Combined score
		score := interactionWeight * recencyFactor * geoRelevance
		articleScores[event.ArticleID] += score
	}
	
	// Convert to slice and sort
	scores := make([]ArticleScore, 0, len(articleScores))
	for articleID, score := range articleScores {
		scores = append(scores, ArticleScore{
			ArticleID: articleID,
			Score:     score,
		})
	}
	
	// Sort by score (descending)
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[i].Score < scores[j].Score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}
	
	// Get top articles
	topCount := limit * 2 // Get extra to account for missing articles
	if topCount > len(scores) {
		topCount = len(scores)
	}
	
	articleIDs := make([]string, topCount)
	for i := 0; i < topCount; i++ {
		articleIDs[i] = scores[i].ArticleID
	}
	
	// Fetch articles
	var articles []models.Article
	if err := database.Where("id IN ?", articleIDs).Find(&articles).Error; err != nil {
		return nil, err
	}
	
	// Sort articles by their trending score
	sortedArticles := make([]models.Article, 0, len(articles))
	for _, score := range scores[:topCount] {
		for _, article := range articles {
			if article.ID == score.ArticleID {
				sortedArticles = append(sortedArticles, article)
				break
			}
		}
	}
	
	// Cache the results
	trendingCache.Set(clusterKey, sortedArticles)
	
	if len(sortedArticles) <= limit {
		return sortedArticles, nil
	}
	return sortedArticles[:limit], nil
}
