package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

//go:embed sql/*.sql
var sqlFiles embed.FS

// Migration represents a database migration.
type Migration struct {
	Version     int
	Description string
	SQL         string
}

const createSchemaTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version INTEGER PRIMARY KEY,
	description TEXT NOT NULL,
	applied_at TEXT NOT NULL DEFAULT (datetime('now'))
)`

// Load reads all embedded SQL files and returns them as migrations.
func Load() ([]Migration, error) {
	entries, err := sqlFiles.ReadDir("sql")
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	var migrations []Migration

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		version, description, err := parseFilename(entry.Name())
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", entry.Name(), err)
		}

		content, err := sqlFiles.ReadFile(filepath.Join("sql", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", entry.Name(), err)
		}

		migrations = append(migrations, Migration{
			Version:     version,
			Description: description,
			SQL:         string(content),
		})
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	// Validate no duplicates
	seen := make(map[int]string)
	for _, m := range migrations {
		if existing, ok := seen[m.Version]; ok {
			return nil, fmt.Errorf("duplicate version %d: %s and %s", m.Version, existing, m.Description)
		}
		seen[m.Version] = m.Description
	}

	return migrations, nil
}

// parseFilename extracts version and description from "NN_description.sql"
func parseFilename(name string) (int, string, error) {
	name = strings.TrimSuffix(name, ".sql")
	parts := strings.SplitN(name, "_", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid format, expected NN_description.sql")
	}

	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("invalid version number: %w", err)
	}

	return version, parts[1], nil
}

// Run executes all pending migrations.
func Run(db *sql.DB) error {
	migrations, err := Load()
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		return nil
	}

	// Ensure schema_migrations exists
	if _, err := db.Exec(createSchemaTable); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	current, err := CurrentVersion(db)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		if m.Version <= current {
			continue
		}

		if err := apply(db, m); err != nil {
			return fmt.Errorf("migration %02d_%s: %w", m.Version, m.Description, err)
		}
	}

	return nil
}

func apply(db *sql.DB, m Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}

	// Use a flag to track if commit succeeded
	committed := false
	defer func() {
		if !committed {
			if rbErr := tx.Rollback(); rbErr != nil {
				// Log rollback error but don't override the original error
				// This is a best-effort cleanup
				_ = rbErr // Rollback failed, but we're already returning an error
			}
		}
	}()

	if _, err := tx.Exec(m.SQL); err != nil {
		return err
	}

	_, err = tx.Exec(
		"INSERT INTO schema_migrations (version, description) VALUES (?, ?)",
		m.Version, m.Description,
	)
	if err != nil {
		return fmt.Errorf("record migration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	committed = true
	return nil
}

// CurrentVersion returns the highest applied migration version.
func CurrentVersion(db *sql.DB) (int, error) {
	// Ensure table exists
	if _, err := db.Exec(createSchemaTable); err != nil {
		return 0, err
	}

	var version sql.NullInt64
	err := db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("get current version: %w", err)
	}
	if !version.Valid {
		return 0, nil
	}
	return int(version.Int64), nil
}

// Pending returns migrations not yet applied.
func Pending(db *sql.DB) ([]Migration, error) {
	migrations, err := Load()
	if err != nil {
		return nil, err
	}

	current, err := CurrentVersion(db)
	if err != nil {
		return nil, err
	}

	var pending []Migration
	for _, m := range migrations {
		if m.Version > current {
			pending = append(pending, m)
		}
	}
	return pending, nil
}
