package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

// Config holds all configuration for our application
type Config struct {
	Port string
	// DatabaseURL string
	JWTSecret   string
	Environment string
}

// Load loads configuratio from environment variables
func Load() *Config {
	// Load .env file if it exists (for local development)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	config := &Config{
		Port:        getEnv("PORT", "8080"),
		JWTSecret:   getEnv("JWT_SECRET", "mysecret"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}
	return config
}

func getEnv(key, defaultValue string) string {
	// Same helper function as in database package
	// You could move this to a shared utils package later
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
