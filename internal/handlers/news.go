package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	readability "github.com/go-shiori/go-readability"
	"github.com/mahigadamsetty/Inshorts-task/internal/config"
	"github.com/mahigadamsetty/Inshorts-task/internal/db"
	"github.com/mahigadamsetty/Inshorts-task/internal/llm"
	"github.com/mahigadamsetty/Inshorts-task/internal/models"
	"github.com/mahigadamsetty/Inshorts-task/internal/services"
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
	category := c.Query("name")
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
	source := c.Query("name")
	limitStr := c.DefaultQuery("limit", "5")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 5
	}

	if source == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name parameter is required"})
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
	searchWords := strings.Split(strings.ToLower(query), " ")
	filteredWords := filterStopWords(searchWords) // Filter stop words
	queryBuilder := database.Model(&models.Article{})

	if len(filteredWords) == 0 {
		filteredWords = searchWords // Fallback to original words if all are stop words
	}

	for _, word := range filteredWords {
		if word != "" {
			searchPattern := "%" + word + "%"
			queryBuilder = queryBuilder.Or("LOWER(title) LIKE ?", searchPattern).Or("LOWER(description) LIKE ?", searchPattern)
		}
	}

	err = queryBuilder.
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

	// Haversine formula in SQL to calculate distance
	// 6371 is the Earth's radius in kilometers
	haversine := fmt.Sprintf(`
		(6371 * acos(cos(radians(%f)) * cos(radians(latitude)) *
		cos(radians(longitude) - radians(%f)) + sin(radians(%f)) *
		sin(radians(latitude))))
	`, lat, lon, lat)

	err = database.
		Select(fmt.Sprintf("*, %s AS distance", haversine)).
		Where(fmt.Sprintf("%s <= ?", haversine), radius).
		Order("distance").
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

	fmt.Printf("result : %+v", result)

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
		searchQuery := result.Query
		if len(result.Entities) > 0 {
			// If entities are found, use them for a more targeted search.
			searchQuery = strings.Join(result.Entities, " ")
		}
		fmt.Println("Executing search with query:", searchQuery) // Debugging line
		searchWords := strings.Split(strings.ToLower(searchQuery), " ")
		filteredWords := filterStopWords(searchWords) // Filter stop words
		queryBuilder := database.Model(&models.Article{})

		if len(filteredWords) == 0 {
			filteredWords = searchWords // Fallback to original words if all are stop words
		}

		for _, word := range filteredWords {
			if word != "" {
				searchPattern := "%" + word + "%"
				queryBuilder = queryBuilder.Or("LOWER(title) LIKE ?", searchPattern).Or("LOWER(description) LIKE ?", searchPattern)
			}
		}

		queryBuilder.Limit(limit * 3).Find(&articles)

		articles = services.RankBySearchRelevance(articles, searchQuery)
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
			var summary string
			var err error

			// Try to get content from URL first
			if articles[i].URL != "" {
				content, err := fetchAndParseURL(articles[i].URL)
				if err == nil && content != "" {
					summary, err = h.llmClient.GenerateSummary(articles[i].Title, content)
				} else if err != nil {
					log.Printf("Failed to fetch or parse URL %s: %v", articles[i].URL, err)
				}
			}

			// Fallback to title and description if URL fetching fails or content is empty
			if summary == "" {
				summary, err = h.llmClient.GenerateSummary(articles[i].Title, articles[i].Description)
			}

			if err == nil {
				articles[i].LLMSummary = summary
				// Optionally save to database
				db.GetDB().Model(&articles[i]).Update("llm_summary", summary)
			} else {
				log.Printf("Failed to generate summary for article %s: %v", articles[i].Title, err)
			}
		}
	}
}

func fetchAndParseURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return "", err
	}
	// Some sites block default user agents
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch URL: status code %d", resp.StatusCode)
	}

	article, err := readability.FromReader(resp.Body, parsedURL)
	if err != nil {
		return "", err
	}

	return article.TextContent, nil
}

var stopWords = map[string]struct{}{
	"a": {}, "about": {}, "above": {}, "after": {}, "again": {}, "against": {}, "all": {}, "am": {}, "an": {}, "and": {}, "any": {}, "are": {}, "as": {}, "at": {},
	"be": {}, "because": {}, "been": {}, "before": {}, "being": {}, "below": {}, "between": {}, "both": {}, "but": {}, "by": {},
	"can": {}, "did": {}, "do": {}, "does": {}, "doing": {}, "don": {}, "down": {}, "during": {},
	"each": {},
	"few":  {}, "for": {}, "from": {}, "further": {},
	"had": {}, "has": {}, "have": {}, "having": {}, "he": {}, "her": {}, "here": {}, "hers": {}, "herself": {}, "him": {}, "himself": {}, "his": {}, "how": {},
	"i": {}, "if": {}, "in": {}, "into": {}, "is": {}, "it": {}, "its": {}, "itself": {},
	"just": {},
	"me":   {}, "more": {}, "most": {}, "my": {}, "myself": {},
	"no": {}, "nor": {}, "not": {},
	"of": {}, "off": {}, "on": {}, "once": {}, "only": {}, "or": {}, "other": {}, "our": {}, "ours": {}, "ourselves": {}, "out": {}, "over": {}, "own": {},
	"s": {}, "same": {}, "she": {}, "should": {}, "so": {}, "some": {}, "such": {},
	"t": {}, "than": {}, "that": {}, "the": {}, "their": {}, "theirs": {}, "them": {}, "themselves": {}, "then": {}, "there": {}, "these": {}, "they": {}, "this": {}, "those": {}, "through": {}, "to": {}, "too": {},
	"under": {}, "until": {}, "up": {},
	"very": {},
	"was":  {}, "we": {}, "were": {}, "what": {}, "when": {}, "where": {}, "which": {}, "while": {}, "who": {}, "whom": {}, "why": {}, "will": {}, "with": {}, "would": {},
	"you": {}, "your": {}, "yours": {}, "yourself": {}, "yourselves": {},
}

func filterStopWords(words []string) []string {
	filtered := make([]string, 0, len(words))
	for _, word := range words {
		if _, found := stopWords[word]; !found {
			filtered = append(filtered, word)
		}
	}
	return filtered
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
