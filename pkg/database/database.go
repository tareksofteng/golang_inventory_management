package database

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens a GORM connection using the given DSN and configures the
// underlying connection pool. It returns the *gorm.DB handle that every
// repository will share.
//
// In Laravel the framework manages the PDO connection pool for you. In Go
// we own the *sql.DB pool explicitly — which is good, because we can tune
// it for production load.
func Connect(dsn string, isProduction bool) (*gorm.DB, error) {
	// In development we log every SQL query (great for learning/debugging).
	// In production we only log slow queries and errors to keep logs clean.
	logLevel := logger.Info
	if isProduction {
		logLevel = logger.Warn
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("database: failed to open connection: %w", err)
	}

	// GORM wraps Go's standard *sql.DB. We reach into it to tune the pool.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("database: failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)                 // max simultaneous open connections
	sqlDB.SetMaxIdleConns(25)                 // connections kept ready in the pool
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // recycle a connection after this long

	// Ping verifies the credentials/host are actually reachable right now,
	// instead of failing later on the first real query.
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database: failed to ping: %w", err)
	}

	return db, nil
}
