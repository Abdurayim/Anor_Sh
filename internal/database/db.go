package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
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
	db.SetMaxOpenConns(1)    // SQLite works best with 1 connection for writes
	db.SetMaxIdleConns(1)    // Keep connection alive
	db.SetConnMaxLifetime(0) // No lifetime limit for SQLite

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

//go:embed migrations/*.sql
var migrationsFS embed.FS

// baseMigration creates the full schema on an empty database.
const baseMigration = "006_complete_schema.sql"

// incrementalMigrations are applied in order on top of the base schema.
var incrementalMigrations = []string{
	"007_fix_parent_students.sql",
	"008_fix_parent_children_view.sql",
	"009_branches.sql",
	"010_super_admin.sql",
}

// Migrate brings the schema up to date. Applied versions are recorded in schema_migrations.
// Databases created before schema_migrations existed are treated as having the base schema.
func Migrate() error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	if _, err := DB.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("failed to create schema_migrations: %w", err)
	}

	var hasSchema bool
	if err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='admins')").Scan(&hasSchema); err != nil {
		return err
	}

	if !hasSchema {
		if err := applyMigration(baseMigration); err != nil {
			return err
		}
	} else if _, err := DB.Exec("INSERT OR IGNORE INTO schema_migrations (version) VALUES (?)", baseMigration); err != nil {
		return err
	}

	for _, name := range incrementalMigrations {
		var applied bool
		if err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)", name).Scan(&applied); err != nil {
			return err
		}
		if applied {
			continue
		}
		if err := applyMigration(name); err != nil {
			return err
		}
	}
	return nil
}

// applyMigration runs one migration file in a transaction. Foreign keys are disabled while it
// runs so tables can be rebuilt (SQLite's 12-step ALTER procedure), and checked afterwards.
func applyMigration(name string) error {
	sqlBytes, err := migrationsFS.ReadFile("migrations/" + name)
	if err != nil {
		return fmt.Errorf("failed to read migration %s: %w", name, err)
	}

	if _, err := DB.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		return err
	}
	defer DB.Exec("PRAGMA foreign_keys = ON")
	// Keep views/triggers from being re-validated while a table is dropped and renamed.
	if _, err := DB.Exec("PRAGMA legacy_alter_table = ON"); err != nil {
		return err
	}
	defer DB.Exec("PRAGMA legacy_alter_table = OFF")

	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(string(sqlBytes)); err != nil {
		return fmt.Errorf("migration %s failed: %w", name, err)
	}
	if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", name); err != nil {
		return err
	}

	rows, err := tx.Query("PRAGMA foreign_key_check")
	if err != nil {
		return err
	}
	violations := rows.Next()
	rows.Close()
	if violations {
		return fmt.Errorf("migration %s left foreign key violations", name)
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	log.Printf("✓ Migration %s applied", name)
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
