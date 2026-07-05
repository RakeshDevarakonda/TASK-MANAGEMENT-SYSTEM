package database

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"
    "task-system/internal/config"

    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
)

var DB *sqlx.DB

// InitDB initializes the connection to the PostgreSQL database and runs migrations.
func InitDB() error {
    var err error
    connStr := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        config.AppConfig.DBHost,
        config.AppConfig.DBPort,
        config.AppConfig.DBUser,
        config.AppConfig.DBPass,
        config.AppConfig.DBName,
        config.AppConfig.DBSSLMode,
    )

    DB, err = sqlx.Connect("postgres", connStr)
    if err != nil {
        return fmt.Errorf("failed to connect to database: %v", err)
    }
    log.Println("Database connection successfully established")

    // Run migrations automatically on start-up.
    if err := runMigrations(); err != nil {
        return fmt.Errorf("migration error: %v", err)
    }
    return nil
}

func runMigrations() error {
    // Determine migration file path. Try executable-relative location first, then fallback to working directory.
    var migrationPath string
    // Try path relative to the executable (useful when binary is in a temporary directory).
    if exePath, exeErr := os.Executable(); exeErr == nil {
        exeDir := filepath.Dir(exePath)
        candidate := filepath.Join(exeDir, "..", "..", "migrations", "schema.sql")
        if _, statErr := os.Stat(candidate); statErr == nil {
            migrationPath = candidate
        }
    }
    // If not found, fallback to the working directory's migrations folder.
    if migrationPath == "" {
        if wd, wdErr := os.Getwd(); wdErr == nil {
            migrationPath = filepath.Join(wd, "migrations", "schema.sql")
        }
    }
    if migrationPath == "" {
        return fmt.Errorf("failed to determine migration file path")
    }
    data, err := os.ReadFile(migrationPath)
    if err != nil {
        return fmt.Errorf("failed to read migration file %s: %w", migrationPath, err)
    }
    // Split statements by semicolon. Trim and ignore empty lines.
    statements := strings.Split(string(data), ";")
    for _, stmt := range statements {
        stmt = strings.TrimSpace(stmt)
        if stmt == "" {
            continue
        }
        // Execute each statement.
        if _, err := DB.Exec(stmt); err != nil {
            return fmt.Errorf("failed to exec migration statement %q: %w", stmt, err)
        }
    }
    log.Println("Database migrations applied successfully")
    return nil
}
