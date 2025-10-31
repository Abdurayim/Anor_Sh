package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Bot       BotConfig
	Database  DatabaseConfig
	Server    ServerConfig
	Admin     AdminConfig
	RateLimit RateLimitConfig
}

type BotConfig struct {
	Token      string
	WebhookURL string
}

type DatabaseConfig struct {
	Path string // Path to SQLite database file
}

type ServerConfig struct {
	Port    string
	GinMode string
}

type AdminConfig struct {
	PhoneNumbers []string
}

type RateLimitConfig struct {
	Requests int
	Duration time.Duration
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load(".env")

	cfg := &Config{
		Bot: BotConfig{
			Token:      getEnv("BOT_TOKEN", ""),
			WebhookURL: getEnv("WEBHOOK_URL", ""),
		},
		Database: DatabaseConfig{
			Path: getEnv("DB_PATH", "parent_bot.db"),
		},
		Server: ServerConfig{
			Port:    getEnv("SERVER_PORT", "8080"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		Admin: AdminConfig{
			PhoneNumbers: parseAdminPhones(getEnv("ADMIN_PHONES", "")),
		},
		RateLimit: RateLimitConfig{
			Requests: 20,
			Duration: 60 * time.Second,
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Bot.Token == "" {
		return fmt.Errorf("BOT_TOKEN is required")
	}

	if len(c.Admin.PhoneNumbers) == 0 {
		return fmt.Errorf("at least one admin phone number is required")
	}

	if len(c.Admin.PhoneNumbers) > 3 {
		return fmt.Errorf("maximum 3 admin phone numbers allowed, got %d", len(c.Admin.PhoneNumbers))
	}

	return nil
}

// GetDBPath returns the SQLite database file path
func (c *DatabaseConfig) GetDBPath() string {
	return c.Path
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// parseAdminPhones parses comma-separated admin phone numbers
func parseAdminPhones(phones string) []string {
	if phones == "" {
		return []string{}
	}

	parts := strings.Split(phones, ",")
	result := make([]string, 0, len(parts))

	for _, phone := range parts {
		trimmed := strings.TrimSpace(phone)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
