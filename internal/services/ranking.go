package services

import (
	"github.com/mahigadamsetty/Inshorts-task/internal/models"
	"github.com/mahigadamsetty/Inshorts-task/internal/utils"
	"sort"
)

type ArticleWithDistance struct {
	Article  *models.Article
	Distance float64
}

type ArticleWithScore struct {
	Article    *models.Article
	MatchScore float64
}

// RankByPublicationDate sorts articles by publication date descending
func RankByPublicationDate(articles []*models.Article) []*models.Article {
	sorted := make([]*models.Article, len(articles))
	copy(sorted, articles)
	
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].PublicationDate.After(sorted[j].PublicationDate)
	})
	
	return sorted
}

// RankByRelevanceScore sorts articles by relevance score descending
func RankByRelevanceScore(articles []*models.Article) []*models.Article {
	sorted := make([]*models.Article, len(articles))
	copy(sorted, articles)
	
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].RelevanceScore > sorted[j].RelevanceScore
	})
	
	return sorted
}

// RankByDistance sorts articles by distance ascending
func RankByDistance(articles []*models.Article, lat, lon float64) []*models.Article {
	withDist := make([]ArticleWithDistance, len(articles))
	
	for i, article := range articles {
		distance := utils.Haversine(lat, lon, article.Latitude, article.Longitude)
		withDist[i] = ArticleWithDistance{
			Article:  article,
			Distance: distance,
		}
	}
	
	sort.Slice(withDist, func(i, j int) bool {
		return withDist[i].Distance < withDist[j].Distance
	})
	
	sorted := make([]*models.Article, len(articles))
	for i, wd := range withDist {
		sorted[i] = wd.Article
	}
	
	return sorted
}

// RankBySearchScore sorts articles by combined relevance (40%) and text match (60%)
func RankBySearchScore(articles []*models.Article, query string) []*models.Article {
	withScore := make([]ArticleWithScore, len(articles))
	
	for i, article := range articles {
		textScore := utils.CalculateTextMatchScore(query, article.Title, article.Description)
		combinedScore := (article.RelevanceScore * 0.4) + (textScore * 0.6)
		withScore[i] = ArticleWithScore{
			Article:    article,
			MatchScore: combinedScore,
		}
	}
	
	sort.Slice(withScore, func(i, j int) bool {
		return withScore[i].MatchScore > withScore[j].MatchScore
	})
	
	sorted := make([]*models.Article, len(articles))
	for i, ws := range withScore {
		sorted[i] = ws.Article
	}
	
	return sorted
}
