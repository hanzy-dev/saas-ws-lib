package db

import (
	"testing"

	wstest "github.com/hanzy-dev/saas-ws-lib/pkg/testkit"
)

func TestMigrateGuard_ForwardOnly(t *testing.T) {
	db := wstest.OpenTestDB(t)
	ctx := wstest.WithTimeout(t, 5_000_000_000)

	if err := EnsureSchemaVersionTable(ctx, db); err != nil {
		t.Fatalf("ensure table: %v", err)
	}

	if err := UpdateSchemaVersion(ctx, db, 1); err != nil {
		t.Fatalf("update version: %v", err)
	}

	if err := EnforceForwardOnly(ctx, db, 0); err == nil {
		t.Fatalf("expected forward-only violation")
	}

	if err := UpdateSchemaVersion(ctx, db, 2); err != nil {
		t.Fatalf("update forward: %v", err)
	}
}
