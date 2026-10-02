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
	Token         string
	WebhookURL    string
	WebhookSecret string // sent by Telegram in X-Telegram-Bot-Api-Secret-Token
}

type DatabaseConfig struct {
	Path string // Path to SQLite database file
}

type ServerConfig struct {
	Port     string
	GinMode  string
	APIToken string // Bearer token for /api/admin/*; API is disabled when empty
}

// Branch codes; they match branches.code in migration 009.
const (
	BranchOlmazor = "olmazor"
	BranchSergeli = "sergeli"
)

// BranchCodes lists branches in display order.
var BranchCodes = []string{BranchOlmazor, BranchSergeli}

type AdminConfig struct {
	// BranchPhones maps branch code -> admin phone numbers of that branch.
	BranchPhones map[string][]string
	// SuperAdminPhone sees statistics of all branches and adds branch admins.
	SuperAdminPhone string
}

// AllPhones returns every configured admin phone.
func (a AdminConfig) AllPhones() []string {
	var out []string
	for _, code := range BranchCodes {
		out = append(out, a.BranchPhones[code]...)
	}
	return out
}

// BranchOf returns the branch code an admin phone is configured for.
func (a AdminConfig) BranchOf(phone string) (string, bool) {
	for _, code := range BranchCodes {
		for _, p := range a.BranchPhones[code] {
			if p == phone {
				return code, true
			}
		}
	}
	return "", false
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
			Token:         getEnv("BOT_TOKEN", ""),
			WebhookURL:    getEnv("WEBHOOK_URL", ""),
			WebhookSecret: getEnv("WEBHOOK_SECRET", ""),
		},
		Database: DatabaseConfig{
			Path: getEnv("DB_PATH", "parent_bot.db"),
		},
		Server: ServerConfig{
			Port:     getEnv("SERVER_PORT", "8080"),
			GinMode:  getEnv("GIN_MODE", "debug"),
			APIToken: getEnv("API_TOKEN", ""),
		},
		Admin: AdminConfig{
			BranchPhones: map[string][]string{
				// ADMIN_PHONES is the pre-branch variable; its phones belong to Olmazor.
				BranchOlmazor: append(parseAdminPhones(getEnv("ADMIN_PHONES_OLMAZOR", "")), parseAdminPhones(getEnv("ADMIN_PHONES", ""))...),
				BranchSergeli: parseAdminPhones(getEnv("ADMIN_PHONES_SERGELI", "")),
			},
			SuperAdminPhone: firstOrEmpty(parseAdminPhones(getEnv("SUPER_ADMIN_PHONE", ""))),
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

	if len(c.Admin.AllPhones()) == 0 && c.Admin.SuperAdminPhone == "" {
		return fmt.Errorf("SUPER_ADMIN_PHONE or at least one branch admin phone (ADMIN_PHONES_OLMAZOR / ADMIN_PHONES_SERGELI) is required")
	}
	if code, ok := c.Admin.BranchOf(c.Admin.SuperAdminPhone); ok {
		return fmt.Errorf("SUPER_ADMIN_PHONE %s is also configured as %s branch admin", c.Admin.SuperAdminPhone, code)
	}

	seen := map[string]string{}
	for _, code := range BranchCodes {
		phones := c.Admin.BranchPhones[code]
		if len(phones) > 3 {
			return fmt.Errorf("maximum 3 admin phone numbers per branch allowed, %s has %d", code, len(phones))
		}
		for _, p := range phones {
			if other, ok := seen[p]; ok && other != code {
				return fmt.Errorf("admin phone %s is configured for both %s and %s", p, other, code)
			}
			seen[p] = code
		}
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

func firstOrEmpty(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// parseAdminPhones parses comma-separated admin phone numbers
func parseAdminPhones(phones string) []string {
	if phones == "" {
		return []string{}
	}

	parts := strings.Split(phones, ",")
	result := make([]string, 0, len(parts))

	for _, phone := range parts {
		trimmed := strings.ReplaceAll(strings.TrimSpace(phone), " ", "")
		if trimmed != "" && !strings.HasPrefix(trimmed, "+") {
			trimmed = "+" + trimmed
		}
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
