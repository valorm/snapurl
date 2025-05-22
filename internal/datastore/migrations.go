package datastore

import (
    "database/sql"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"
)

// RunMigrations applies all pending .sql files in the migrations directory
func RunMigrations(db *sql.DB) error {
    const migrationsDir = "migrations"

    entries, err := os.ReadDir(migrationsDir)
    if err != nil {
        return fmt.Errorf("read migrations dir: %w", err)
    }

    // Collect and sort .sql filenames
    var files []string
    for _, e := range entries {
        if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
            files = append(files, e.Name())
        }
    }
    sort.Strings(files)

    // Execute each migration in order
    for _, name := range files {
        path := filepath.Join(migrationsDir, name)
        sqlBytes, err := os.ReadFile(path)
        if err != nil {
            return fmt.Errorf("read migration %s: %w", name, err)
        }
        if _, err := db.Exec(string(sqlBytes)); err != nil {
            return fmt.Errorf("exec migration %s: %w", name, err)
        }
    }

    return nil
}
