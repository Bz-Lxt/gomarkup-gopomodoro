package store

import (
	"context"
	"os"
	"testing"
)

// TestMigrateReleasesAdvisoryLock verifies that the advisory lock is released
// after Migrate returns. Regression test for the rolling-deploy deadlock where
// the lock leaked back into the pool on an idle session and blocked the next
// instance from acquiring it.
//
// The test needs a real Postgres connection. If no DATABASE_URL is available
// it skips (the CI e2e/docker-compose stack provides one).
func TestMigrateReleasesAdvisoryLock(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://gopomo:gopomo_dev@127.0.0.1:35174/gopomodoro?sslmode=disable"
	}

	ctx := context.Background()
	db, err := Open(ctx, dsn)
	if err != nil {
		t.Skipf("db not available, skipping: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.Migrate(ctx, "../../migrations"); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// After Migrate returns, the advisory lock MUST be free. A second
	// acquisition should succeed immediately; if the first instance still
	// held it (leaked into the pool) the rolling-deploy deadlock recurs.
	var got bool
	if err := db.SQL.QueryRowContext(ctx,
		"SELECT pg_try_advisory_lock($1)", advisoryLockKey).Scan(&got); err != nil {
		t.Fatalf("query pg_try_advisory_lock: %v", err)
	}
	if !got {
		t.Fatal("advisory lock still held after Migrate returned; " +
			"it leaked into the connection pool")
	}
	// Release the lock we just took so the DB is left clean.
	var released bool
	_ = db.SQL.QueryRowContext(ctx,
		"SELECT pg_advisory_unlock($1)", advisoryLockKey).Scan(&released)
}
