package store

import (
	"database/sql"
	"sync"

	"github.com/footprint-tools/cli/internal/log"
	"github.com/footprint-tools/cli/internal/store/migrations"
)

var (
	db          *sql.DB
	once        sync.Once
	openError   error
	singletonMu sync.RWMutex
)

// Open opens the database and runs any pending migrations.
// Uses singleton pattern - subsequent calls return the same connection.
//
// Deprecated: Use store.New() instead for proper dependency injection.
func Open(path string) (*sql.DB, error) {
	once.Do(func() {
		log.Debug("store: opening database at %s", path)

		var err error
		conn, err := sql.Open("sqlite3", path)
		if err != nil {
			singletonMu.Lock()
			openError = err
			singletonMu.Unlock()
			log.Error("store: failed to open database: %v", err)
			return
		}

		if err = conn.Ping(); err != nil {
			_ = conn.Close()
			singletonMu.Lock()
			openError = err
			singletonMu.Unlock()
			log.Error("store: failed to ping database: %v", err)
			return
		}

		if err = configureSQLite(conn); err != nil {
			_ = conn.Close()
			singletonMu.Lock()
			openError = err
			singletonMu.Unlock()
			log.Error("store: failed to configure database: %v", err)
			return
		}

		setDBPermissions(path)

		if err = migrations.Run(conn); err != nil {
			_ = conn.Close()
			singletonMu.Lock()
			openError = err
			singletonMu.Unlock()
			log.Error("store: migrations failed: %v", err)
			return
		}

		singletonMu.Lock()
		db = conn
		singletonMu.Unlock()
		log.Debug("store: database ready")
	})

	singletonMu.RLock()
	defer singletonMu.RUnlock()
	return db, openError
}

// OpenFresh opens a new database connection without singleton.
// Used for testing with in-memory databases.
//
// Deprecated: Use store.New() instead.
func OpenFresh(path string) (*sql.DB, error) {
	s, err := New(path)
	if err != nil {
		return nil, err
	}
	return s.DB(), nil
}

// ResetSingleton resets the singleton state. Only for testing.
//
// Deprecated: Use store.New() instead which doesn't use singletons.
func ResetSingleton() {
	singletonMu.Lock()
	defer singletonMu.Unlock()
	once = sync.Once{}
	db = nil
	openError = nil
}
