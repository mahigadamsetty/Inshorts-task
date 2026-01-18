package db

import (
	"fmt"
	"log"

	"github.com/mahigadamsetty/Inshorts-task/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init initializes the database connection and runs migrations
func Init(databaseURL string) error {
	var err error
	
	// Open database connection
	DB, err = gorm.Open(sqlite.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations
	if err := DB.AutoMigrate(&models.Article{}, &models.Event{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
