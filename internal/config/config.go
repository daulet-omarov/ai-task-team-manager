package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string
	AppUrl  string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	JWTSecret string

	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	AppBaseURL   string
}

func Load() *Config {

	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found")
	}

	cfg := &Config{
		AppPort: getEnv("APP_PORT", "8080"),
		AppUrl:  getEnv("APP_URL", "localhost:8080"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "myapp"),

		JWTSecret: getEnv("JWT_SECRET", "secret"),

		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", ""),
		AppBaseURL:   getEnv("APP_BASE_URL", "http://localhost:7777"),
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
