package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"parent-bot/internal/config"

	_ "github.com/mattn/go-sqlite3"
)

// DB is the global database connection
var DB *sql.DB

// Connect establishes database connection with connection pooling
func Connect(cfg *config.DatabaseConfig) error {
	// SQLite uses a file path instead of DSN
	dbPath := cfg.GetDBPath()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for SQLite
	db.SetMaxOpenConns(1)                  // SQLite works best with 1 connection for writes
	db.SetMaxIdleConns(1)                  // Keep connection alive
	db.SetConnMaxLifetime(0)               // No lifetime limit for SQLite

	// Enable foreign keys and WAL mode for better performance
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	_, err = db.Exec("PRAGMA journal_mode = WAL")
	if err != nil {
		return fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// RunMigrations executes SQL migration files
func RunMigrations(migrationPath string) error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	sqlBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	_, err = DB.Exec(string(sqlBytes))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}

// HealthCheck checks if database is reachable
func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return DB.PingContext(ctx)
}
