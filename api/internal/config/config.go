package config

import (
	"log"
	"os"
)

// Config holds all configuration for the application.
type Config struct {
	AppPort  string
	MongoURI string
	MongoDB  string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	cfg := &Config{
		AppPort:  getEnv("APP_PORT", "8080"),
		MongoURI: getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:  getEnv("MONGO_DB_NAME", "rev_saas"),
	}

	log.Printf("Config loaded: port=%s, mongo_db=%s", cfg.AppPort, cfg.MongoDB)

	return cfg
}

// getEnv retrieves an environment variable or returns a fallback value.
func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

