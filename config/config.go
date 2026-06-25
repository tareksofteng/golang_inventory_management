package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds every configuration value the application needs.
// We group related values into sub-structs (e.g. DB) so call sites
// read naturally: cfg.DB.Host instead of cfg.DBHost.
type Config struct {
	AppEnv  string
	AppPort string
	DB      DBConfig
	JWT     JWTConfig
	Seed    SeedConfig
}

// SeedConfig holds the credentials for the super-admin created on first run.
type SeedConfig struct {
	AdminEmail    string
	AdminPassword string
}

// JWTConfig holds token signing settings. The secret MUST be overridden in
// production via env — never ship the default.
type JWTConfig struct {
	Secret           string
	AccessTTLMinutes int
	RefreshTTLHours  int
}

// DBConfig holds the database connection settings.
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

// Load reads the .env file (if present) and builds a Config.
// In production the .env file usually does NOT exist; real environment
// variables are injected by the server/container instead. That is why a
// missing .env is only a warning, not a fatal error.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("config: no .env file found, falling back to OS environment variables")
	}

	return &Config{
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "8080"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "127.0.0.1"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "inventory_db"),
		},
		JWT: JWTConfig{
			Secret:           getEnv("JWT_SECRET", "change-me-in-production"),
			AccessTTLMinutes: getEnvInt("JWT_ACCESS_TTL_MIN", 15),
			RefreshTTLHours:  getEnvInt("JWT_REFRESH_TTL_HOURS", 168), // 7 days
		},
		Seed: SeedConfig{
			AdminEmail:    getEnv("SEED_ADMIN_EMAIL", "admin@inventory.test"),
			AdminPassword: getEnv("SEED_ADMIN_PASSWORD", "Admin@123"),
		},
	}
}

// getEnvInt is like getEnv but parses an integer, falling back on parse errors.
func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		if n, err := strconv.Atoi(value); err == nil {
			return n
		}
	}
	return fallback
}

// DSN builds the MySQL connection string GORM expects.
// parseTime=True  -> scan DATETIME columns into time.Time
// loc=Local       -> store/read timestamps in the server's local timezone
// charset=utf8mb4 -> full unicode (emoji, Bangla) support
func (db DBConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		db.User, db.Password, db.Host, db.Port, db.Name,
	)
}

// getEnv returns the environment variable for key, or fallback when the
// variable is unset or empty. A small helper keeps Load() clean.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
