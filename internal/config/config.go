package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string
	AppUrl  string
	TLSCert string
	TLSKey  string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret string

	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	AppBaseURL   string

	AllowedOrigins string

	S3Endpoint  string
	S3Region    string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string
	S3PublicURL string
}

func Load() *Config {

	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found")
	}

	cfg := &Config{
		AppPort: getEnv("APP_PORT", "8080"),
		AppUrl:  getEnv("APP_URL", "localhost:8080"),
		TLSCert: getEnv("TLS_CERT", ""),
		TLSKey:  getEnv("TLS_KEY", ""),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "myapp"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		JWTSecret: getEnv("JWT_SECRET", "secret"),

		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", ""),
		AppBaseURL:   getEnv("APP_BASE_URL", "http://localhost:8080"),

		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:5173"),

		S3Endpoint:  getEnv("S3_ENDPOINT", ""),
		S3Region:    getEnv("S3_REGION", ""),
		S3AccessKey: getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey: getEnv("S3_SECRET_KEY", ""),
		S3Bucket:    getEnv("S3_BUCKET", ""),
		S3PublicURL: getEnv("S3_PUBLIC_URL", ""),
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
