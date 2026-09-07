package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	Dsn                 string
	JwtAccessSecret     string
	JwtRefreshSecret    string
	JwtAccessExpiry     string
	JwtRefreshExpiry    string
	CloudinaryCloudName string
	CloudinaryApiKey    string
	CloudinaryApiSecret string
	AppEnv              string // "development" | "staging" | "production"

	// SMTP / Gmail App Password config
	SMTPFromName             string
	SMTPFromAddress          string
	SMTPAppPassword          string
	SMTPOTPExpirationMinutes int // Default: 10
	FirebaseProjectID        string

	// Admin Settings
	AdminEmail    string
	AdminPassword string

	// Media Settings
	MaxUploadSizeMB int
}

func LoadEnv() *Config {
	// Load .env file for local development.
	// In production (Docker), environment variables are injected by the runtime,
	// so the absence of a .env file is expected and not an error.
	_ = godotenv.Load()

	return &Config{
		Port:                     getEnvWithDefault("PORT", "5525"),
		Dsn:                      os.Getenv("DSN"),
		JwtAccessSecret:          os.Getenv("JWT_ACCESS_TOKEN_SECRET"),
		JwtRefreshSecret:         os.Getenv("JWT_REFRESH_TOKEN_SECRET"),
		JwtAccessExpiry:          os.Getenv("JWT_ACCESS_TOKEN_EXPIRY"),
		JwtRefreshExpiry:         os.Getenv("JWT_REFRESH_TOKEN_EXPIRY"),
		CloudinaryCloudName:      os.Getenv("CLOUDINARY_CLOUD_NAME"),
		CloudinaryApiKey:         os.Getenv("CLOUDINARY_API_KEY"),
		CloudinaryApiSecret:      os.Getenv("CLOUDINARY_API_SECRET"),
		AppEnv:                   getEnvWithDefault("APP_ENV", "development"),
		SMTPFromName:             getEnvWithDefault("SMTP_FROM_NAME", "Alter App"),
		SMTPFromAddress:          os.Getenv("SMTP_FROM_ADDRESS"),
		SMTPAppPassword:          os.Getenv("SMTP_APP_PASSWORD"),
		SMTPOTPExpirationMinutes: getEnvInt("SMTP_OTP_EXPIRATION_MINUTES", 10),
		FirebaseProjectID:        os.Getenv("FIREBASE_PROJECT_ID"),
		AdminEmail:               os.Getenv("ADMIN_EMAIL"),
		AdminPassword:            os.Getenv("ADMIN_PASSWORD"),
		MaxUploadSizeMB:          getEnvInt("MAX_UPLOAD_SIZE_MB", 200),
	}
}

func getEnvWithDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}
