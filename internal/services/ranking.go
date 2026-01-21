package services

import (
	"sort"
	"strings"

	"github.com/mahigadamsetty/Inshorts-task/internal/models"
	"github.com/mahigadamsetty/Inshorts-task/internal/utils"
)

// ArticleWithScore wraps an article with a computed score
type ArticleWithScore struct {
	Article models.Article
	Score   float64
}

// RankByPublicationDate ranks articles by publication date (newest first)
func RankByPublicationDate(articles []models.Article) []models.Article {
	// Already sorted by database query
	return articles
}

// RankByRelevanceScore ranks articles by relevance score (highest first)
func RankByRelevanceScore(articles []models.Article) []models.Article {
	// Already sorted by database query
	return articles
}

// RankByDistance ranks articles by distance from a location (nearest first)
func RankByDistance(articles []models.Article, lat, lon float64) []models.Article {
	scored := make([]ArticleWithScore, len(articles))

	for i, article := range articles {
		distance := utils.HaversineDistance(lat, lon, article.Latitude, article.Longitude)
		scored[i] = ArticleWithScore{
			Article: article,
			Score:   distance,
		}
	}

	// Sort by distance (ascending) using built-in sort
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score < scored[j].Score
	})

	result := make([]models.Article, len(scored))
	for i, s := range scored {
		result[i] = s.Article
	}

	return result
}

// RankBySearchRelevance ranks articles by how well they match the search query.
// It calculates a dynamic score based on keyword matches in the title and description.
func RankBySearchRelevance(articles []models.Article, query string) []models.Article {
	scored := make([]ArticleWithScore, len(articles))
	queryWords := strings.Fields(strings.ToLower(query))

	// Filter out stop words from the query to focus on meaningful terms
	queryWords = filterStopWords(queryWords)

	for i, article := range articles {
		score := calculateTextMatchScore(article, queryWords)
		scored[i] = ArticleWithScore{
			Article: article,
			Score:   score,
		}
	}

	// Sort by the dynamically calculated score (descending)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	result := make([]models.Article, len(scored))
	for i, s := range scored {
		result[i] = s.Article
	}

	return result
}

// calculateTextMatchScore computes a text match score based on query terms.
// Matches in the title are weighted more heavily than matches in the description.
func calculateTextMatchScore(article models.Article, queryWords []string) float64 {
	if len(queryWords) == 0 {
		return 0
	}

	titleLower := strings.ToLower(article.Title)
	descLower := strings.ToLower(article.Description)

	var score float64
	titleWeight := 3.0 // Title matches are 3x more important
	descWeight := 1.0

	for _, word := range queryWords {
		if strings.Contains(titleLower, word) {
			score += titleWeight
		}
		if strings.Contains(descLower, word) {
			score += descWeight
		}
	}

	// Normalize the score by the number of query words to avoid favoring longer queries
	return score / float64(len(queryWords))
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
