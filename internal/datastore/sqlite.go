package datastore

import (
    "database/sql"
    "fmt"
    "os"
    "path/filepath"

    _ "github.com/mattn/go-sqlite3"
)

// OpenDB opens (or creates) the SQLite file at `path` and runs migrations.
func OpenDB(path string) (*sql.DB, error) {
    // Ensure parent directory exists
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0o755); err != nil {
        return nil, fmt.Errorf("create data dir: %w", err)
    }

    db, err := sql.Open("sqlite3", path)
    if err != nil {
        return nil, fmt.Errorf("open sqlite db: %w", err)
    }

    // Run all migrations
    if err := RunMigrations(db); err != nil {
        db.Close()
        return nil, fmt.Errorf("run migrations: %w", err)
    }

    return db, nil
}
