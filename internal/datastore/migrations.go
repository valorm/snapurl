package datastore

import (
    "database/sql"
    "embed"
    "fmt"
    "io/fs"
    "path"
    "sort"
    "strings"
)

// Embed all SQL files under migrations/
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations applies all pending .sql files in order, ignoring
// duplicate-column errors so it can be run idempotently.
func RunMigrations(db *sql.DB) error {
    const migrationsDir = "migrations"

    // Read the embedded migrations directory
    entries, err := fs.ReadDir(migrationsFS, migrationsDir)
    if err != nil {
        return fmt.Errorf("read migrations dir: %w", err)
    }

    // Sort so migrations run in lexicographical order
    sort.Slice(entries, func(i, j int) bool {
        return entries[i].Name() < entries[j].Name()
    })

    // Execute each .sql file
    for _, entry := range entries {
        name := entry.Name()
        if !strings.HasSuffix(name, ".sql") {
            continue
        }

        // Build the embedded path (always with forward slashes)
        p := path.Join(migrationsDir, name)

        sqlBytes, err := migrationsFS.ReadFile(p)
        if err != nil {
            return fmt.Errorf("read migration %s: %w", name, err)
        }

        if _, err := db.Exec(string(sqlBytes)); err != nil {
            // If it's an "ALTER TABLE ... ADD COLUMN" that's already been applied, skip it
            if strings.Contains(err.Error(), "duplicate column name") {
                continue
            }
            return fmt.Errorf("exec migration %s: %w", name, err)
        }
    }

    return nil
}
