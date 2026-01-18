package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mahigadamsetty/Inshorts-task/internal/config"
	"github.com/mahigadamsetty/Inshorts-task/internal/db"
	"github.com/mahigadamsetty/Inshorts-task/internal/llm"
	"github.com/mahigadamsetty/Inshorts-task/internal/models"
	"github.com/mahigadamsetty/Inshorts-task/internal/services"
	"github.com/mahigadamsetty/Inshorts-task/internal/utils"
)

type NewsHandler struct {
	llmClient *llm.Client
	config    *config.Config
}

func NewNewsHandler(cfg *config.Config) *NewsHandler {
	return &NewsHandler{
		llmClient: llm.NewClient(cfg.OpenAIAPIKey, cfg.LLMModel),
		config:    cfg,
	}
}

type Response struct {
	Articles []models.Article `json:"articles"`
	Meta     Meta             `json:"meta"`
}

type Meta struct {
	Count    int    `json:"count"`
	Limit    int    `json:"limit"`
	Endpoint string `json:"endpoint"`
	Query    string `json:"query,omitempty"`
}

// GetByCategory handles /category endpoint
func (h *NewsHandler) GetByCategory(c *gin.Context) {
	category := c.Query("category")
	limitStr := c.DefaultQuery("limit", "5")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}
	
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category parameter is required"})
		return
	}
	
	database := db.GetDB()
	var articles []models.Article
	
	// Search for articles containing the category (case-insensitive)
	err = database.
		Where("LOWER(category) LIKE ?", "%"+strings.ToLower(category)+"%").
		Order("publication_date DESC").
		Limit(limit).
		Find(&articles).Error
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch articles"})
		return
	}
	
	// Enrich with summaries
	h.enrichWithSummaries(articles)
	
	c.JSON(http.StatusOK, Response{
		Articles: articles,
		Meta: Meta{
			Count:    len(articles),
			Limit:    limit,
			Endpoint: "category",
			Query:    category,
		},
	})
}

// GetBySource handles /source endpoint
func (h *NewsHandler) GetBySource(c *gin.Context) {
	source := c.Query("source")
	limitStr := c.DefaultQuery("limit", "5")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}
	
	if source == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source parameter is required"})
		return
	}
	
	database := db.GetDB()
	var articles []models.Article
	
	err = database.
		Where("LOWER(source_name) = ?", strings.ToLower(source)).
		Order("publication_date DESC").
		Limit(limit).
		Find(&articles).Error
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch articles"})
		return
	}
	
	// Enrich with summaries
	h.enrichWithSummaries(articles)
	
	c.JSON(http.StatusOK, Response{
		Articles: articles,
		Meta: Meta{
			Count:    len(articles),
			Limit:    limit,
			Endpoint: "source",
			Query:    source,
		},
	})
}

// GetByScore handles /score endpoint
func (h *NewsHandler) GetByScore(c *gin.Context) {
	minStr := c.Query("min")
	limitStr := c.DefaultQuery("limit", "5")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}
	
	minScore := 0.0
	if minStr != "" {
		minScore, err = strconv.ParseFloat(minStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid min score"})
			return
		}
	}
	
	database := db.GetDB()
	var articles []models.Article
	
	err = database.
		Where("relevance_score >= ?", minScore).
		Order("relevance_score DESC").
		Limit(limit).
		Find(&articles).Error
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch articles"})
		return
	}
	
	// Enrich with summaries
	h.enrichWithSummaries(articles)
	
	c.JSON(http.StatusOK, Response{
		Articles: articles,
		Meta: Meta{
			Count:    len(articles),
			Limit:    limit,
			Endpoint: "score",
			Query:    minStr,
		},
	})
}

// Search handles /search endpoint
func (h *NewsHandler) Search(c *gin.Context) {
	query := c.Query("query")
	limitStr := c.DefaultQuery("limit", "5")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}
	
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter is required"})
		return
	}
	
	database := db.GetDB()
	var articles []models.Article
	
	// Search in title and description
	searchPattern := "%" + strings.ToLower(query) + "%"
	err = database.
		Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchPattern, searchPattern).
		Limit(limit * 3). // Get more to rank properly
		Find(&articles).Error
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch articles"})
		return
	}
	
	// Rank by search relevance
	articles = services.RankBySearchRelevance(articles, query)
	
	// Limit results
	if len(articles) > limit {
		articles = articles[:limit]
	}
	
	// Enrich with summaries
	h.enrichWithSummaries(articles)
	
	c.JSON(http.StatusOK, Response{
		Articles: articles,
		Meta: Meta{
			Count:    len(articles),
			Limit:    limit,
			Endpoint: "search",
			Query:    query,
		},
	})
}

// GetNearby handles /nearby endpoint
func (h *NewsHandler) GetNearby(c *gin.Context) {
	latStr := c.Query("lat")
	lonStr := c.Query("lon")
	radiusStr := c.DefaultQuery("radius", "10")
	limitStr := c.DefaultQuery("limit", "5")
	
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid latitude"})
		return
	}
	
	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid longitude"})
		return
	}
	
	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil || radius <= 0 {
		radius = 10
	}
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}
	
	database := db.GetDB()
	var articles []models.Article
	
	// Fetch all articles (could be optimized with spatial index)
	err = database.Find(&articles).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch articles"})
		return
	}
	
	// Rank by distance and filter by radius
	articles = services.RankByDistance(articles, lat, lon)
	
	// Filter by radius and limit
	filtered := []models.Article{}
	for _, article := range articles {
		distance := utils.HaversineDistance(lat, lon, article.Latitude, article.Longitude)
		if distance <= radius {
			filtered = append(filtered, article)
			if len(filtered) >= limit {
				break
			}
		}
	}
	
	// Enrich with summaries
	h.enrichWithSummaries(filtered)
	
	c.JSON(http.StatusOK, Response{
		Articles: filtered,
		Meta: Meta{
			Count:    len(filtered),
			Limit:    limit,
			Endpoint: "nearby",
			Query:    latStr + "," + lonStr,
		},
	})
}

// GetTrending handles /trending endpoint
func (h *NewsHandler) GetTrending(c *gin.Context) {
	latStr := c.Query("lat")
	lonStr := c.Query("lon")
	limitStr := c.DefaultQuery("limit", "5")
	
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid latitude"})
		return
	}
	
	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid longitude"})
		return
	}
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}
	
	articles, err := services.GetTrendingArticles(lat, lon, limit, h.config.LocationClusterDegrees)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trending articles"})
		return
	}
	
	// Enrich with summaries
	h.enrichWithSummaries(articles)
	
	c.JSON(http.StatusOK, Response{
		Articles: articles,
		Meta: Meta{
			Count:    len(articles),
			Limit:    limit,
			Endpoint: "trending",
			Query:    latStr + "," + lonStr,
		},
	})
}

// Query handles /query endpoint (LLM-powered)
func (h *NewsHandler) Query(c *gin.Context) {
	query := c.Query("query")
	latStr := c.Query("lat")
	lonStr := c.Query("lon")
	limitStr := c.DefaultQuery("limit", "5")
	
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter is required"})
		return
	}
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}
	
	// Extract intent and entities using LLM
	result, err := h.llmClient.ExtractIntentAndEntities(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process query"})
		return
	}
	
	// Dispatch to appropriate endpoint based on intent
	var articles []models.Article
	endpoint := result.Intent
	
	database := db.GetDB()
	
	switch result.Intent {
	case llm.IntentCategory:
		// Extract category from query or entities
		category := extractCategory(query, result.Entities)
		if category != "" {
			database.
				Where("LOWER(category) LIKE ?", "%"+strings.ToLower(category)+"%").
				Order("publication_date DESC").
				Limit(limit).
				Find(&articles)
		}
		
	case llm.IntentSource:
		// Extract source from query or entities
		source := extractSource(query, result.Entities)
		if source != "" {
			database.
				Where("LOWER(source_name) LIKE ?", "%"+strings.ToLower(source)+"%").
				Order("publication_date DESC").
				Limit(limit).
				Find(&articles)
		}
		
	case llm.IntentScore:
		database.
			Where("relevance_score >= ?", 0.7).
			Order("relevance_score DESC").
			Limit(limit).
			Find(&articles)
		
	case llm.IntentNearby:
		if latStr != "" && lonStr != "" {
			lat, _ := strconv.ParseFloat(latStr, 64)
			lon, _ := strconv.ParseFloat(lonStr, 64)
			
			database.Find(&articles)
			articles = services.RankByDistance(articles, lat, lon)
			if len(articles) > limit {
				articles = articles[:limit]
			}
		}
		
	default: // IntentSearch
		searchPattern := "%" + strings.ToLower(result.Query) + "%"
		database.
			Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchPattern, searchPattern).
			Limit(limit * 3).
			Find(&articles)
		
		articles = services.RankBySearchRelevance(articles, result.Query)
		if len(articles) > limit {
			articles = articles[:limit]
		}
	}
	
	// Enrich with summaries
	h.enrichWithSummaries(articles)
	
	c.JSON(http.StatusOK, Response{
		Articles: articles,
		Meta: Meta{
			Count:    len(articles),
			Limit:    limit,
			Endpoint: endpoint,
			Query:    query,
		},
	})
}

// enrichWithSummaries adds LLM-generated summaries to articles
func (h *NewsHandler) enrichWithSummaries(articles []models.Article) {
	for i := range articles {
		if articles[i].LLMSummary == "" {
			summary, err := h.llmClient.GenerateSummary(articles[i].Title, articles[i].Description)
			if err == nil {
				articles[i].LLMSummary = summary
				// Optionally save to database
				db.GetDB().Model(&articles[i]).Update("llm_summary", summary)
			}
		}
	}
}

// Helper functions
func extractCategory(query string, entities []string) string {
	categories := []string{
		"technology", "tech", "sports", "business", "entertainment",
		"science", "health", "politics", "world", "national", "general",
	}
	
	lowerQuery := strings.ToLower(query)
	for _, cat := range categories {
		if strings.Contains(lowerQuery, cat) {
			return cat
		}
	}
	
	return ""
}

func extractSource(query string, entities []string) string {
	// Common news sources
	sources := []string{
		"new york times", "washington post", "cnn", "bbc", "reuters",
		"associated press", "guardian", "wall street journal",
	}
	
	lowerQuery := strings.ToLower(query)
	for _, source := range sources {
		if strings.Contains(lowerQuery, source) {
			return source
		}
	}
	
	// Check entities for potential source names
	for _, entity := range entities {
		if len(entity) > 3 {
			return entity
		}
	}
	
	return ""
}
