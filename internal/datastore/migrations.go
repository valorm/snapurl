package datastore

import (
    "database/sql"
    "embed"
    "fmt"
    "io/fs"
    "path/filepath"
    "sort"
    "strings"
)

// Embed all SQL files under migrations/
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations applies all pending .sql files in order.
func RunMigrations(db *sql.DB) error {
    const migrationsDir = "migrations"

    entries, err := fs.ReadDir(migrationsFS, migrationsDir)
    if err != nil {
        return fmt.Errorf("read migrations dir: %w", err)
    }

    // Sort by filename so versions run in order
    sort.Slice(entries, func(i, j int) bool {
        return entries[i].Name() < entries[j].Name()
    })

    for _, entry := range entries {
        name := entry.Name()
        if !strings.HasSuffix(name, ".sql") {
            continue
        }

        path := filepath.Join(migrationsDir, name)
        sqlBytes, err := migrationsFS.ReadFile(path)
        if err != nil {
            return fmt.Errorf("read migration %s: %w", name, err)
        }

        if _, err := db.Exec(string(sqlBytes)); err != nil {
            return fmt.Errorf("exec migration %s: %w", name, err)
        }
    }

    return nil
}
