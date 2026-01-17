package services

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/mahigadamsetty/Inshorts-task/internal/db"
	"github.com/mahigadamsetty/Inshorts-task/internal/models"
	"github.com/mahigadamsetty/Inshorts-task/internal/utils"
)

type TrendingCache struct {
	articles  []*models.Article
	expiresAt time.Time
}

type TrendingService struct {
	cache              map[string]*TrendingCache
	cacheMutex         sync.RWMutex
	cacheTTL           time.Duration
	clusterDegrees     float64
	simulationRunning  bool
	simulationMutex    sync.Mutex
}

func NewTrendingService(cacheTTLSeconds int, clusterDegrees float64) *TrendingService {
	return &TrendingService{
		cache:          make(map[string]*TrendingCache),
		cacheTTL:       time.Duration(cacheTTLSeconds) * time.Second,
		clusterDegrees: clusterDegrees,
	}
}

// StartEventSimulation starts simulating user activity events
func (ts *TrendingService) StartEventSimulation(ctx context.Context) {
	ts.simulationMutex.Lock()
	if ts.simulationRunning {
		ts.simulationMutex.Unlock()
		return
	}
	ts.simulationRunning = true
	ts.simulationMutex.Unlock()

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ts.simulateEvents()
			}
		}
	}()
}

func (ts *TrendingService) simulateEvents() {
	// Get random articles from database
	var articles []models.Article
	db.GetDB().Limit(100).Find(&articles)
	
	if len(articles) == 0 {
		return
	}

	// Generate 5-20 random events
	numEvents := rand.Intn(16) + 5
	
	for i := 0; i < numEvents; i++ {
		article := articles[rand.Intn(len(articles))]
		
		// Random event type (60% views, 40% clicks)
		eventType := models.EventTypeView
		if rand.Float64() < 0.4 {
			eventType = models.EventTypeClick
		}
		
		// Use article location with some random variation
		lat := article.Latitude + (rand.Float64()-0.5)*2.0
		lon := article.Longitude + (rand.Float64()-0.5)*2.0
		
		event := models.Event{
			ArticleID: article.ID,
			EventType: eventType,
			Latitude:  lat,
			Longitude: lon,
			CreatedAt: time.Now(),
		}
		
		db.GetDB().Create(&event)
	}
}

// GetTrendingArticles retrieves trending articles for a location
func (ts *TrendingService) GetTrendingArticles(lat, lon float64, limit int) ([]*models.Article, error) {
	clusterKey := utils.GetLocationClusterKey(lat, lon, ts.clusterDegrees)
	
	// Check cache
	ts.cacheMutex.RLock()
	cached, exists := ts.cache[clusterKey]
	ts.cacheMutex.RUnlock()
	
	if exists && time.Now().Before(cached.expiresAt) {
		if len(cached.articles) <= limit {
			return cached.articles, nil
		}
		return cached.articles[:limit], nil
	}
	
	// Compute trending articles
	articles, err := ts.computeTrending(lat, lon, limit*3) // Get more for better ranking
	if err != nil {
		return nil, err
	}
	
	// Cache results
	ts.cacheMutex.Lock()
	ts.cache[clusterKey] = &TrendingCache{
		articles:  articles,
		expiresAt: time.Now().Add(ts.cacheTTL),
	}
	ts.cacheMutex.Unlock()
	
	if len(articles) <= limit {
		return articles, nil
	}
	return articles[:limit], nil
}

type articleScore struct {
	article *models.Article
	score   float64
}

func (ts *TrendingService) computeTrending(lat, lon float64, limit int) ([]*models.Article, error) {
	// Get recent events (last hour)
	var events []models.Event
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	
	err := db.GetDB().Where("created_at > ?", oneHourAgo).Find(&events).Error
	if err != nil {
		return nil, err
	}
	
	// Get unique article IDs
	articleEventCounts := make(map[string][]models.Event)
	for _, event := range events {
		articleEventCounts[event.ArticleID] = append(articleEventCounts[event.ArticleID], event)
	}
	
	if len(articleEventCounts) == 0 {
		// No recent events, return recent articles by publication date
		var articles []*models.Article
		err := db.GetDB().Order("publication_date DESC").Limit(limit).Find(&articles).Error
		return articles, err
	}
	
	// Get articles
	articleIDs := make([]string, 0, len(articleEventCounts))
	for id := range articleEventCounts {
		articleIDs = append(articleIDs, id)
	}
	
	var articles []models.Article
	err = db.GetDB().Where("id IN ?", articleIDs).Find(&articles).Error
	if err != nil {
		return nil, err
	}
	
	// Calculate trending scores
	scores := make([]articleScore, 0, len(articles))
	now := time.Now()
	
	for _, article := range articles {
		events := articleEventCounts[article.ID]
		
		// Calculate interaction score
		interactionScore := 0.0
		for _, event := range events {
			weight := 1.0
			if event.EventType == models.EventTypeClick {
				weight = 2.0 // Clicks count more
			}
			
			// Recency decay (exponential)
			age := now.Sub(event.CreatedAt).Hours()
			recencyFactor := math.Exp(-age / 12.0) // Decay over 12 hours
			
			// Geographical relevance
			distance := utils.Haversine(lat, lon, event.Latitude, event.Longitude)
			geoFactor := 1.0 / (1.0 + distance/100.0) // Decay with distance
			
			interactionScore += weight * recencyFactor * geoFactor
		}
		
		scores = append(scores, articleScore{
			article: &article,
			score:   interactionScore,
		})
	}
	
	// Sort by score
	for i := 0; i < len(scores); i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].score > scores[i].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}
	
	// Convert to article slice
	result := make([]*models.Article, 0, limit)
	for i := 0; i < len(scores) && i < limit; i++ {
		result = append(result, scores[i].article)
	}
	
	return result, nil
}

// ClearCache clears the trending cache
func (ts *TrendingService) ClearCache() {
	ts.cacheMutex.Lock()
	defer ts.cacheMutex.Unlock()
	ts.cache = make(map[string]*TrendingCache)
}
