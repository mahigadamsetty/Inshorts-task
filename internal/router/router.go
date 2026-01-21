package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/mahigadamsetty/Inshorts-task/internal/config"
	"github.com/mahigadamsetty/Inshorts-task/internal/handlers"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)
	
	r := gin.Default()
	
	// CORS middleware - allow all for demo
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	
	// Initialize handlers
	newsHandler := handlers.NewNewsHandler(cfg)
	
	// API v1 routes
	v1 := r.Group("/api/v1/news")
	{
		v1.GET("/category", newsHandler.GetByCategory)
		v1.GET("/source", newsHandler.GetBySource)
		v1.GET("/score", newsHandler.GetByScore)
		v1.GET("/search", newsHandler.Search)
		v1.GET("/nearby", newsHandler.GetNearby)
		v1.GET("/trending", newsHandler.GetTrending)
		v1.GET("/query", newsHandler.Query)
	}
	
	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	
	return r
}
