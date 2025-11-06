package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	DBHost       string
	DBPort       string
	DBName       string
	DBUser       string
	DBPassword   string
	DatabaseURL  string
	KafkaBroker  string
	RedisURL     string
	NodeEnv      string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		Port:         getEnv("PORT", "3001"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBName:       getEnv("DB_NAME", "four_in_a_row"),
		DBUser:       getEnv("DB_USER", "postgres"),
		DBPassword:   getEnv("DB_PASSWORD", ""),
		DatabaseURL:  getEnv("DATABASE_URL", ""),
		KafkaBroker:  getEnv("KAFKA_BROKER", ""),
		RedisURL:     getEnv("REDIS_URL", ""),
		NodeEnv:      getEnv("NODE_ENV", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}