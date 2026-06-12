package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort          string
	SSOAPIURL        string
	ClientURL        string
	GoogleAPIKey     string
	GoogleCalendarID string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

var cfg *Config

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg = &Config{
		AppPort:          getEnv("APP_PORT", "8081"),
		SSOAPIURL:        getEnv("SSO_API_URL", "http://localhost:8080"),
		ClientURL:        getEnv("CLIENT_URL", "http://localhost:3001"),
		GoogleAPIKey:     getEnv("GOOGLE_API_KEY", ""),
		GoogleCalendarID: getEnv("GOOGLE_CALENDAR_ID", ""),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "laci_v3"),
	}

	return cfg
}

func Get() *Config {
	if cfg == nil {
		return Load()
	}
	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
