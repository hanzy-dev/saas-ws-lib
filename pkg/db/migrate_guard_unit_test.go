package db

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMigrateGuard_NilDB(t *testing.T) {
	t.Parallel()

	if err := EnsureSchemaVersionTable(context.Background(), nil); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := CurrentSchemaVersion(context.Background(), nil); err == nil {
		t.Fatalf("expected error")
	}
	if err := UpdateSchemaVersion(context.Background(), nil, 1); err == nil {
		t.Fatalf("expected error")
	}
}

func TestEnforceForwardOnly_InvalidVersion(t *testing.T) {
	t.Parallel()
	db := openFakeDB("mg_invalid", &fakeState{})
	defer db.Close()

	if err := EnforceForwardOnly(context.Background(), db, -1); err == nil {
		t.Fatalf("expected error")
	}
}

func TestSchemaVersionFlow_Unit(t *testing.T) {
	t.Parallel()

	st := &fakeState{version: 0}
	db := openFakeDB("mg_flow", st)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := EnsureSchemaVersionTable(ctx, db); err != nil {
		t.Fatalf("ensure: %v", err)
	}

	v, err := CurrentSchemaVersion(ctx, db)
	if err != nil || v != 0 {
		t.Fatalf("current=%d err=%v", v, err)
	}

	if err := UpdateSchemaVersion(ctx, db, 2); err != nil {
		t.Fatalf("update: %v", err)
	}

	v2, _ := CurrentSchemaVersion(ctx, db)
	if v2 != 2 {
		t.Fatalf("expected 2, got %d", v2)
	}

	// forward-only violation
	if err := EnforceForwardOnly(ctx, db, 1); err == nil {
		t.Fatalf("expected violation")
	}

	// simulate query failure
	st.mu.Lock()
	st.queryErr = errors.New("qfail")
	st.mu.Unlock()
	if _, err := CurrentSchemaVersion(ctx, db); err == nil {
		t.Fatalf("expected error")
	}
}
