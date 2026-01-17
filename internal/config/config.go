package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL            string
	OpenAIAPIKey           string
	LLMModel               string
	TrendingCacheTTL       int // seconds
	LocationClusterDegrees float64
	Port                   string
}

func Load() *Config {
	return &Config{
		DatabaseURL:            getEnv("DATABASE_URL", "news.db"),
		OpenAIAPIKey:           getEnv("OPENAI_API_KEY", ""),
		LLMModel:               getEnv("LLM_MODEL", "gpt-4o-mini"),
		TrendingCacheTTL:       getEnvInt("TRENDING_CACHE_TTL", 300), // 5 minutes default
		LocationClusterDegrees: getEnvFloat("LOCATION_CLUSTER_DEGREES", 0.5),
		Port:                   getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}
