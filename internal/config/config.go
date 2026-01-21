package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL            string
	OpenAIAPIKey           string
	LLMModel               string
	TrendingCacheTTL       int
	LocationClusterDegrees float64
	Port                   string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file, will use environment variables if set")
	}
	return &Config{
		DatabaseURL:            getEnv("DATABASE_URL", "news.db"),
		OpenAIAPIKey:           getEnv("OPENAI_API_KEY", ""),
		LLMModel:               getEnv("LLM_MODEL", "gpt-4o-mini"),
		TrendingCacheTTL:       getEnvAsInt("TRENDING_CACHE_TTL", 300),
		LocationClusterDegrees: getEnvAsFloat("LOCATION_CLUSTER_DEGREES", 0.5),
		Port:                   getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}
