package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mahigadamsetty/Inshorts-task/internal/db"
	"github.com/mahigadamsetty/Inshorts-task/internal/llm"
	"github.com/mahigadamsetty/Inshorts-task/internal/models"
	"github.com/mahigadamsetty/Inshorts-task/internal/services"
	"github.com/mahigadamsetty/Inshorts-task/internal/utils"
)

type NewsHandler struct {
	llmClient       *llm.Client
	trendingService *services.TrendingService
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

func NewNewsHandler(llmClient *llm.Client, trendingService *services.TrendingService) *NewsHandler {
	return &NewsHandler{
		llmClient:       llmClient,
		trendingService: trendingService,
	}
}

// GetByCategory handles /category endpoint
func (h *NewsHandler) GetByCategory(c *gin.Context) {
	category := c.Query("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category parameter is required"})
		return
	}

	limit := h.getLimit(c)

	var articles []*models.Article
	err := db.GetDB().Where("category LIKE ?", "%"+category+"%").Find(&articles).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// Rank by publication date
	articles = services.RankByPublicationDate(articles)

	// Limit results
	if len(articles) > limit {
		articles = articles[:limit]
	}

	// Enrich with summaries
	h.enrichWithSummaries(c.Request.Context(), articles)

	h.sendResponse(c, articles, limit, "category", category)
}

// GetBySource handles /source endpoint
func (h *NewsHandler) GetBySource(c *gin.Context) {
	source := c.Query("source")
	if source == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source parameter is required"})
		return
	}

	limit := h.getLimit(c)

	var articles []*models.Article
	err := db.GetDB().Where("source_name LIKE ?", "%"+source+"%").Find(&articles).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// Rank by publication date
	articles = services.RankByPublicationDate(articles)

	// Limit results
	if len(articles) > limit {
		articles = articles[:limit]
	}

	// Enrich with summaries
	h.enrichWithSummaries(c.Request.Context(), articles)

	h.sendResponse(c, articles, limit, "source", source)
}

// GetByScore handles /score endpoint
func (h *NewsHandler) GetByScore(c *gin.Context) {
	minStr := c.Query("min")
	if minStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "min parameter is required"})
		return
	}

	min, err := strconv.ParseFloat(minStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid min parameter"})
		return
	}

	limit := h.getLimit(c)

	var articles []*models.Article
	err = db.GetDB().Where("relevance_score >= ?", min).Find(&articles).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// Rank by relevance score
	articles = services.RankByRelevanceScore(articles)

	// Limit results
	if len(articles) > limit {
		articles = articles[:limit]
	}

	// Enrich with summaries
	h.enrichWithSummaries(c.Request.Context(), articles)

	h.sendResponse(c, articles, limit, "score", fmt.Sprintf("min=%.2f", min))
}

// Search handles /search endpoint
func (h *NewsHandler) Search(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter is required"})
		return
	}

	limit := h.getLimit(c)

	// Get all articles (or a reasonable subset)
	var articles []*models.Article
	err := db.GetDB().Find(&articles).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// Rank by search score (40% relevance + 60% text match)
	articles = services.RankBySearchScore(articles, query)

	// Limit results
	if len(articles) > limit {
		articles = articles[:limit]
	}

	// Enrich with summaries
	h.enrichWithSummaries(c.Request.Context(), articles)

	h.sendResponse(c, articles, limit, "search", query)
}

// GetNearby handles /nearby endpoint
func (h *NewsHandler) GetNearby(c *gin.Context) {
	latStr := c.Query("lat")
	lonStr := c.Query("lon")
	radiusStr := c.Query("radius")

	if latStr == "" || lonStr == "" || radiusStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat, lon, and radius parameters are required"})
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lat parameter"})
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lon parameter"})
		return
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid radius parameter"})
		return
	}

	limit := h.getLimit(c)

	// Get all articles
	var allArticles []*models.Article
	err = db.GetDB().Find(&allArticles).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// Filter by radius
	var articles []*models.Article
	for _, article := range allArticles {
		distance := utils.Haversine(lat, lon, article.Latitude, article.Longitude)
		if distance <= radius {
			articles = append(articles, article)
		}
	}

	// Rank by distance
	articles = services.RankByDistance(articles, lat, lon)

	// Limit results
	if len(articles) > limit {
		articles = articles[:limit]
	}

	// Enrich with summaries
	h.enrichWithSummaries(c.Request.Context(), articles)

	query := fmt.Sprintf("lat=%.4f, lon=%.4f, radius=%.1fkm", lat, lon, radius)
	h.sendResponse(c, articles, limit, "nearby", query)
}

// GetTrending handles /trending endpoint
func (h *NewsHandler) GetTrending(c *gin.Context) {
	latStr := c.Query("lat")
	lonStr := c.Query("lon")

	if latStr == "" || lonStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat and lon parameters are required"})
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lat parameter"})
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lon parameter"})
		return
	}

	limit := h.getLimit(c)

	articles, err := h.trendingService.GetTrendingArticles(lat, lon, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get trending articles"})
		return
	}

	// Enrich with summaries
	h.enrichWithSummaries(c.Request.Context(), articles)

	query := fmt.Sprintf("lat=%.4f, lon=%.4f", lat, lon)
	h.sendResponse(c, articles, limit, "trending", query)
}

// Query handles /query endpoint (LLM-powered)
func (h *NewsHandler) Query(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter is required"})
		return
	}

	// Extract intent and entities
	result, err := h.llmClient.ExtractEntitiesAndIntent(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process query"})
		return
	}

	// Dispatch to appropriate endpoint based on intent
	switch result.Intent {
	case llm.IntentCategory:
		if category, ok := result.Entities["category"]; ok {
			c.Request.URL.RawQuery = fmt.Sprintf("category=%s&limit=%d", category, h.getLimit(c))
			h.GetByCategory(c)
			return
		}

	case llm.IntentSource:
		if source, ok := result.Entities["source"]; ok {
			c.Request.URL.RawQuery = fmt.Sprintf("source=%s&limit=%d", source, h.getLimit(c))
			h.GetBySource(c)
			return
		}

	case llm.IntentScore:
		c.Request.URL.RawQuery = fmt.Sprintf("min=0.7&limit=%d", h.getLimit(c))
		h.GetByScore(c)
		return

	case llm.IntentNearby:
		// Use provided lat/lon or default
		lat := c.Query("lat")
		lon := c.Query("lon")
		if lat == "" || lon == "" {
			// Default location if not provided
			lat = "37.4220"
			lon = "-122.0840"
		}
		c.Request.URL.RawQuery = fmt.Sprintf("lat=%s&lon=%s&radius=50&limit=%d", lat, lon, h.getLimit(c))
		h.GetNearby(c)
		return

	case llm.IntentSearch:
		// Use search endpoint
		searchQuery := query
		if keywords, ok := result.Entities["keywords"]; ok {
			searchQuery = keywords
		}
		c.Request.URL.RawQuery = fmt.Sprintf("query=%s&limit=%d", searchQuery, h.getLimit(c))
		h.Search(c)
		return
	}

	// Default to search if intent unclear
	c.Request.URL.RawQuery = fmt.Sprintf("query=%s&limit=%d", query, h.getLimit(c))
	h.Search(c)
}

func (h *NewsHandler) getLimit(c *gin.Context) int {
	limitStr := c.Query("limit")
	if limitStr == "" {
		return 5
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		return 5
	}

	if limit > 100 {
		return 100
	}

	return limit
}

func (h *NewsHandler) enrichWithSummaries(ctx context.Context, articles []*models.Article) {
	for _, article := range articles {
		// Only generate summary if not already present
		if article.LLMSummary == "" {
			summary, err := h.llmClient.Summarize(ctx, article.Title, article.Description)
			if err == nil {
				article.LLMSummary = summary
				// Save to database to cache it
				db.GetDB().Model(article).Update("llm_summary", summary)
			}
		}
	}
}

func (h *NewsHandler) sendResponse(c *gin.Context, articles []*models.Article, limit int, endpoint, query string) {
	// Convert to non-pointer slice for JSON response
	result := make([]models.Article, len(articles))
	for i, article := range articles {
		result[i] = *article
	}

	response := Response{
		Articles: result,
		Meta: Meta{
			Count:    len(result),
			Limit:    limit,
			Endpoint: endpoint,
			Query:    query,
		},
	}

	c.JSON(http.StatusOK, response)
}
