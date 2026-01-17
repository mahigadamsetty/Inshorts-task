package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/mahigadamsetty/Inshorts-task/internal/handlers"
)

func Setup(newsHandler *handlers.NewsHandler) *gin.Engine {
	r := gin.Default()

	// CORS middleware - allow all for demo
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	// API routes
	api := r.Group("/api/v1/news")
	{
		api.GET("/category", newsHandler.GetByCategory)
		api.GET("/source", newsHandler.GetBySource)
		api.GET("/score", newsHandler.GetByScore)
		api.GET("/search", newsHandler.Search)
		api.GET("/nearby", newsHandler.GetNearby)
		api.GET("/trending", newsHandler.GetTrending)
		api.GET("/query", newsHandler.Query)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
