package services

import (
	"fmt"
	"math"
	"sort"
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

// ArticleScore represents an article with its trending score
type ArticleScore struct {
	ArticleID string
	Score     float64
}

// GetTrendingArticles calculates and returns trending articles based on user events
func GetTrendingArticles(lat, lon float64, limit int, clusterDegrees float64) ([]models.Article, error) {
	// Use a geospatial cluster key for caching
	clusterKey := getClusterKey(lat, lon, clusterDegrees)

	// Check cache first
	if articles, found := trendingCache.Get(clusterKey); found {
		if len(articles) > limit {
			return articles[:limit], nil
		}
		return articles, nil
	}

	// --- If not in cache, calculate trending scores ---
	database := db.GetDB()

	// 1. Fetch recent events (e.g., last 24 hours)
	var recentEvents []models.Event
	err := database.Where("timestamp > ?", time.Now().Add(-24*time.Hour)).Find(&recentEvents).Error
	if err != nil {
		return nil, err
	}

	if len(recentEvents) == 0 {
		// If no recent events, return empty or a fallback (e.g., latest articles)
		return []models.Article{}, nil
	}

	// 2. Calculate trending score for each article
	articleScores := make(map[string]float64)
	articleIDs := make(map[string]bool)

	for _, event := range recentEvents {
		score := calculateEventScore(event, lat, lon)
		articleScores[event.ArticleID] += score
		articleIDs[event.ArticleID] = true
	}

	// 3. Get the article details for the trending articles
	var ids []string
	for id := range articleIDs {
		ids = append(ids, id)
	}

	var articles []models.Article
	err = database.Where("id IN ?", ids).Find(&articles).Error
	if err != nil {
		return nil, err
	}

	// 4. Attach scores and sort
	for i := range articles {
		articles[i].TrendingScore = articleScores[articles[i].ID]
	}

	// Sort articles by trending score in descending order
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].TrendingScore > articles[j].TrendingScore
	})

	// Limit the results
	if len(articles) > limit {
		articles = articles[:limit]
	}

	trendingCache.Set(clusterKey, articles)

	return articles, nil
}

// calculateEventScore computes a score for a single user event
func calculateEventScore(event models.Event, userLat, userLon float64) float64 {
	// Base score for event type
	baseScore := 1.0 // View
	if event.EventType == "click" {
		baseScore = 3.0 // Clicks are more valuable
	}

	// Time decay factor (events from the last hour are most valuable)
	hoursAgo := time.Since(event.Timestamp).Hours()
	timeDecay := math.Exp(-0.1 * hoursAgo) // Exponential decay

	// Location proximity factor
	distance := utils.HaversineDistance(userLat, userLon, event.Latitude, event.Longitude)
	locationFactor := math.Exp(-0.05 * distance) // Closer events get higher score

	return baseScore * timeDecay * locationFactor
}

// getClusterKey creates a string key for a geographic cluster.
func getClusterKey(lat, lon, clusterDegrees float64) string {
	latCluster := math.Round(lat/clusterDegrees) * clusterDegrees
	lonCluster := math.Round(lon/clusterDegrees) * clusterDegrees
	return fmt.Sprintf("%.2f,%.2f", latCluster, lonCluster)
}
