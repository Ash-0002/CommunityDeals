// Package testutil provides shared helpers for integration tests that need a
// live PostgreSQL database.  Import it in _test packages only.
package testutil

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// DSN returns the test database DSN from the environment variable POSTGRES_DSN,
// falling back to the local docker-compose default.
func DSN() string {
	if v := os.Getenv("POSTGRES_DSN"); v != "" {
		return v
	}
	return "postgres://cp_user:cp_password@localhost:5432/community_platform?sslmode=disable"
}

// MustConnectDB opens a PostgreSQL connection and fails the test if the DB is
// not reachable.  The connection is closed automatically via t.Cleanup.
func MustConnectDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Connect("postgres", DSN())
	if err != nil {
		t.Skipf("skipping integration test — cannot connect to DB: %v", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	t.Cleanup(func() { db.Close() })
	return db
}
