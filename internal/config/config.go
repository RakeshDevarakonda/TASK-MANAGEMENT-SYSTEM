package config

import (
    "os"
    "fmt"
)

type Config struct {
    DBHost               string
    DBPort               string
    DBUser               string
    DBPass               string
    DBName               string
    DBSSLMode            string
    JWTSecret            string
    EncryptionKey        string // derived from PNV; used for AES encryption
}

var AppConfig *Config

// LoadConfig loads environment configurations into the global AppConfig variable.
// It panics if a required high‑entropy JWT secret is missing – this ensures the
// application cannot start without a proper secret.
func LoadConfig() {
    AppConfig = &Config{
        DBHost:               getEnvOrDefault("DB_HOST", "localhost"),
        DBPort:               getEnvOrDefault("DB_PORT", "5432"),
        DBUser:               getEnvOrDefault("DB_USER", "postgres"),
        DBPass:               getEnvOrDefault("DB_PASS", ""),
        DBName:               getEnvOrDefault("DB_NAME", "tasksystems"),
        DBSSLMode:            getEnvOrDefault("DB_SSLMODE", "disable"),
        JWTSecret:            getEnvOrPanic("JWT_SECRET"),
        EncryptionKey:        getEnvOrPanic("ENCRYPTION_KEY"),
    }
}

func getEnvOrDefault(key, defaultValue string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return defaultValue
}

func getEnvOrPanic(key string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    panic(fmt.Sprintf("environment variable %s must be set and contain a high‑entropy secret", key))
}
