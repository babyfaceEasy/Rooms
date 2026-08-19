package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config aggregates all application configuration values.
type Config struct {
	App struct {
		Name     string
		Port     string
		Env      string
		LogLevel string
	}
	Mongo struct {
		URI            string
		Database       string
		ConnectTimeout time.Duration
	}
	S3 struct {
		EndpointURL     string
		Region          string
		Bucket          string
		AccessKey       string
		SecretKey       string
		UseSSL          bool
		PresignedExpiry time.Duration
	}
	JWT struct {
		Secret          string
		AccessTokenTTL  time.Duration
		RefreshTokenTTL time.Duration
	}
	SendGrid struct {
		APIKey                  string
		SenderEmail             string
		Enabled                 bool
		VerificationTemplateID  string
		PasswordResetTemplateID string
	}
	RateLimit struct {
		GlobalMax   int
		GlobalWindow time.Duration
		AuthMax     int
		AuthWindow  time.Duration
	}
}

// Load reads configuration from environment variables and applies sensible defaults.
func Load() (Config, error) {
	cfg := Config{}

	cfg.App.Name = getEnv("APP_NAME", "temp_backend")
	cfg.App.Port = getEnv("PORT", "8080")
	cfg.App.Env = getEnv("APP_ENV", "development")
	cfg.App.LogLevel = getEnv("LOG_LEVEL", "info")

	cfg.Mongo.URI = getEnv("MONGO_URI", "mongodb://localhost:27017/temp_backend")
	cfg.Mongo.Database = getEnv("MONGO_DATABASE", "temp_backend")
	cfg.Mongo.ConnectTimeout = parseDuration(getEnv("MONGO_CONNECT_TIMEOUT", "10s"))

	cfg.S3.EndpointURL = getEnv("S3_ENDPOINT_URL", "http://localhost:9000")
	cfg.S3.Region = getEnv("S3_REGION", "us-east-1")
	cfg.S3.Bucket = getEnv("S3_BUCKET", "temp-bucket")
	cfg.S3.AccessKey = getEnv("S3_ACCESS_KEY", "minioadmin")
	cfg.S3.SecretKey = getEnv("S3_SECRET_KEY", "minioadmin")
	cfg.S3.UseSSL = parseBool(getEnv("S3_USE_SSL", "false"))
	cfg.S3.PresignedExpiry = parseDuration(getEnv("S3_PRESIGNED_EXPIRY", "15m"))

	cfg.JWT.Secret = getEnv("JWT_SECRET", "")
	cfg.JWT.AccessTokenTTL = parseDuration(getEnv("ACCESS_TOKEN_TTL", "1h"))
	cfg.JWT.RefreshTokenTTL = parseDuration(getEnv("REFRESH_TOKEN_TTL", "168h"))

	cfg.RateLimit.GlobalMax = getInt("RATE_LIMIT_GLOBAL_MAX", 100)
	cfg.RateLimit.GlobalWindow = parseDuration(getEnv("RATE_LIMIT_GLOBAL_WINDOW", "1m"))
	cfg.RateLimit.AuthMax = getInt("RATE_LIMIT_AUTH_MAX", 20)
	cfg.RateLimit.AuthWindow = parseDuration(getEnv("RATE_LIMIT_AUTH_WINDOW", "1m"))

	cfg.SendGrid.APIKey = getEnv("SENDGRID_API_KEY", "")
	cfg.SendGrid.SenderEmail = getEnv("SENDGRID_SENDER_EMAIL", "noreply@tempbackend.com")
	cfg.SendGrid.Enabled = parseBool(getEnv("SENDGRID_ENABLED", "true"))
	cfg.SendGrid.VerificationTemplateID = getEnv("SENDGRID_VERIFICATION_TEMPLATE_ID", "")
	cfg.SendGrid.PasswordResetTemplateID = getEnv("SENDGRID_PASSWORD_RESET_TEMPLATE_ID", "")

	if cfg.Mongo.URI == "" {
		return cfg, fmt.Errorf("MONGO_URI is required")
	}
	if cfg.S3.EndpointURL == "" {
		return cfg, fmt.Errorf("S3_ENDPOINT_URL is required")
	}
	if cfg.S3.Bucket == "" {
		return cfg, fmt.Errorf("S3_BUCKET is required")
	}
	if cfg.JWT.Secret == "" {
		return cfg, fmt.Errorf("JWT_SECRET is required")
	}
	if len(cfg.JWT.Secret) < 32 {
		return cfg, fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func parseBool(value string) bool {
	ok, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return ok
}

func parseDuration(value string) time.Duration {
	if value == "" {
		return 0
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return 0
	}
	return d
}
