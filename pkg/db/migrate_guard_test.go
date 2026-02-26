//go:build integration
// +build integration

package db_test

import (
	"testing"
	"time"

	wsdb "github.com/hanzy-dev/saas-ws-lib/pkg/db"
	wstest "github.com/hanzy-dev/saas-ws-lib/pkg/testkit"
)

func TestMigrateGuard_ForwardOnly(t *testing.T) {
	db := wstest.OpenTestDB(t)
	ctx := wstest.WithTimeout(t, 5*time.Second)

	if err := wsdb.EnsureSchemaVersionTable(ctx, db); err != nil {
		t.Fatalf("ensure table: %v", err)
	}

	if err := wsdb.UpdateSchemaVersion(ctx, db, 1); err != nil {
		t.Fatalf("update version: %v", err)
	}

	if err := wsdb.EnforceForwardOnly(ctx, db, 0); err == nil {
		t.Fatalf("expected forward-only violation")
	}

	if err := wsdb.UpdateSchemaVersion(ctx, db, 2); err != nil {
		t.Fatalf("update forward: %v", err)
	}
}
