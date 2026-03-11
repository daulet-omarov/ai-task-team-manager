package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	JWTSecret string
}

func Load() *Config {

	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found")
	}

	cfg := &Config{
		AppPort: getEnv("APP_PORT", "8080"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "myapp"),

		JWTSecret: getEnv("JWT_SECRET", "secret"),
	}

	return cfg
}

func getEnv(key, fallback string) string {

	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
