package services

import (
	"math"
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
	
	// Sort by distance (ascending)
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[i].Score > scored[j].Score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}
	
	result := make([]models.Article, len(scored))
	for i, s := range scored {
		result[i] = s.Article
	}
	
	return result
}

// RankBySearchRelevance ranks articles by combined relevance score and text match
// 40% relevance_score + 60% text match score
func RankBySearchRelevance(articles []models.Article, query string) []models.Article {
	scored := make([]ArticleWithScore, len(articles))
	queryLower := strings.ToLower(query)
	queryWords := strings.Fields(queryLower)
	
	for i, article := range articles {
		textScore := calculateTextMatchScore(article, queryWords)
		combinedScore := (article.RelevanceScore * 0.4) + (textScore * 0.6)
		
		scored[i] = ArticleWithScore{
			Article: article,
			Score:   combinedScore,
		}
	}
	
	// Sort by combined score (descending)
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[i].Score < scored[j].Score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}
	
	result := make([]models.Article, len(scored))
	for i, s := range scored {
		result[i] = s.Article
	}
	
	return result
}

// calculateTextMatchScore computes a text match score based on query terms
func calculateTextMatchScore(article models.Article, queryWords []string) float64 {
	titleLower := strings.ToLower(article.Title)
	descLower := strings.ToLower(article.Description)
	
	if len(queryWords) == 0 {
		return 0
	}
	
	matchCount := 0
	titleMatches := 0
	
	for _, word := range queryWords {
		if len(word) < 2 {
			continue
		}
		
		if strings.Contains(titleLower, word) {
			titleMatches++
			matchCount++
		} else if strings.Contains(descLower, word) {
			matchCount++
		}
	}
	
	// Title matches are weighted more heavily
	score := float64(titleMatches)*0.6 + float64(matchCount-titleMatches)*0.4
	maxScore := float64(len(queryWords))
	
	if maxScore == 0 {
		return 0
	}
	
	return math.Min(score/maxScore, 1.0)
}
