package database

import (
	"fmt"
	"time"

	"github.com/community-platform/backend/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// NewPostgres creates and validates a PostgreSQL connection pool.
func NewPostgres(cfg config.DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}

	return db, nil
}

// MustNewPostgres is like NewPostgres but panics on error. Use during startup.
func MustNewPostgres(cfg config.DatabaseConfig) *sqlx.DB {
	db, err := NewPostgres(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to postgres: %v", err))
	}
	return db
}

// HealthCheck verifies the database connection is alive.
func HealthCheck(db *sqlx.DB) error {
	ctx := db.DB // underlying *sql.DB
	_ = ctx
	if err := db.Ping(); err != nil {
		return fmt.Errorf("postgres health check failed: %w", err)
	}
	return nil
}

// WithRetry tries to connect to PostgreSQL with exponential back-off.
// Useful when the app and DB start simultaneously in Docker Compose.
func WithRetry(cfg config.DatabaseConfig, maxAttempts int) (*sqlx.DB, error) {
	var (
		db  *sqlx.DB
		err error
	)

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		db, err = NewPostgres(cfg)
		if err == nil {
			return db, nil
		}

		wait := time.Duration(attempt) * 2 * time.Second
		time.Sleep(wait)
	}

	return nil, fmt.Errorf("could not connect to postgres after %d attempts: %w", maxAttempts, err)
}
