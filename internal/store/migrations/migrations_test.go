package migrations_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/footprint-tools/cli/internal/store/migrations"
)

func TestLoad(t *testing.T) {
	all, err := migrations.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if len(all) < 2 {
		t.Fatalf("expected at least 2 migrations, got %d", len(all))
	}

	// Verify strictly increasing order
	for i := 1; i < len(all); i++ {
		if all[i].Version <= all[i-1].Version {
			t.Errorf("migration %d (v%d) not after %d (v%d)",
				i, all[i].Version, i-1, all[i-1].Version)
		}
	}
}

func TestRunIdempotent(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = db.Close() }()

	// First run
	if err := migrations.Run(db); err != nil {
		t.Fatalf("first run: %v", err)
	}

	v1, err := migrations.CurrentVersion(db)
	if err != nil {
		t.Fatalf("get version: %v", err)
	}

	// Second run - should be idempotent
	if err := migrations.Run(db); err != nil {
		t.Fatalf("second run: %v", err)
	}

	v2, err := migrations.CurrentVersion(db)
	if err != nil {
		t.Fatalf("get version: %v", err)
	}

	if v1 != v2 {
		t.Errorf("version changed: %d -> %d", v1, v2)
	}
}

func TestPending(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = db.Close() }()

	all, _ := migrations.Load()

	// Before run: all pending
	pending, err := migrations.Pending(db)
	if err != nil {
		t.Fatalf("pending before: %v", err)
	}
	if len(pending) != len(all) {
		t.Errorf("expected %d pending, got %d", len(all), len(pending))
	}

	// After run: none pending
	if err := migrations.Run(db); err != nil {
		t.Fatalf("run: %v", err)
	}
	pending, _ = migrations.Pending(db)
	if len(pending) != 0 {
		t.Errorf("expected 0 pending, got %d", len(pending))
	}
}

func TestTablesCreated(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := migrations.Run(db); err != nil {
		t.Fatalf("run: %v", err)
	}

	tables := []string{"schema_migrations", "event_status", "event_source", "repo_events"}
	for _, table := range tables {
		var name string
		err := db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		if err == sql.ErrNoRows {
			t.Errorf("table %s not created", table)
		} else if err != nil {
			t.Errorf("check %s: %v", table, err)
		}
	}
}

func TestRedundantColumnsRemoved(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := migrations.Run(db); err != nil {
		t.Fatalf("run: %v", err)
	}

	// Verify commit_message and author columns were removed
	for _, col := range []string{"commit_message", "author"} {
		var count int
		err = db.QueryRow(`
			SELECT COUNT(*) FROM pragma_table_info('repo_events') WHERE name = ?
		`, col).Scan(&count)
		if err != nil {
			t.Fatalf("check %s: %v", col, err)
		}
		if count != 0 {
			t.Errorf("%s column should not exist", col)
		}
	}
}
