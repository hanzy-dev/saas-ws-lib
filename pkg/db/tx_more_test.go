package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

func TestWithTx_BeginError(t *testing.T) {
	t.Parallel()

	st := &fakeState{beginErr: errors.New("begin fail")}
	db := openFakeDB("tx_begin_fail", st)
	defer db.Close()

	err := WithTx(context.Background(), db, TxOptions{}, func(ctx context.Context, tx *sql.Tx) error { return nil })
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestWithTx_CommitError(t *testing.T) {
	t.Parallel()

	st := &fakeState{commitErr: errors.New("commit fail")}
	db := openFakeDB("tx_commit_fail", st)
	defer db.Close()

	err := WithTx(context.Background(), db, TxOptions{}, func(ctx context.Context, tx *sql.Tx) error { return nil })
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestWithTxDefault_CallsWithTx(t *testing.T) {
	t.Parallel()

	st := &fakeState{}
	db := openFakeDB("tx_default", st)
	defer db.Close()

	err := WithTxDefault(context.Background(), db, func(ctx context.Context, tx *sql.Tx) error { return nil })
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}
