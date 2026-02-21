package testkit

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	wsdb "github.com/hanzy-dev/saas-ws-lib/pkg/db"
)

const EnvTestDBDSN = "TEST_DB_DSN"

func OpenTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn, ok := os.LookupEnv(EnvTestDBDSN)
	if !ok || dsn == "" {
		t.Skip("TEST_DB_DSN is not set; skipping DB integration test")
	}

	db, err := wsdb.Open(wsdb.Config{
		DSN:             dsn,
		MaxOpenConns:    10,
		MaxIdleConns:    10,
		ConnMaxIdleTime: 2 * time.Minute,
		ConnMaxLifetime: 10 * time.Minute,
		PingTimeout:     5 * time.Second,
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })
	return db
}

func WithTimeout(t *testing.T, d time.Duration) context.Context {
	t.Helper()
	if d <= 0 {
		d = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), d)
	t.Cleanup(cancel)
	return ctx
}
