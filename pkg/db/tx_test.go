package db

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
)

func TestWithTx_NilDB(t *testing.T) {
	t.Parallel()
	err := WithTx(context.Background(), nil, TxOptions{}, func(ctx context.Context, tx *sql.Tx) error { return nil })
	if !errors.Is(err, sql.ErrConnDone) {
		t.Fatalf("expected sql.ErrConnDone, got %v", err)
	}
}

func TestWithTx_NilFn(t *testing.T) {
	t.Parallel()
	db := openFakeDB("tx_nil_fn", &fakeState{})
	defer db.Close()

	err := WithTx(context.Background(), db, TxOptions{}, nil)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestWithTx_CommitsOnSuccess(t *testing.T) {
	t.Parallel()
	st := &fakeState{}
	db := openFakeDB("tx_commit", st)
	defer db.Close()

	err := WithTx(context.Background(), db, TxOptions{}, func(ctx context.Context, tx *sql.Tx) error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	st.mu.Lock()
	defer st.mu.Unlock()
	if st.commits != 1 || st.rollbacks != 0 {
		t.Fatalf("commits=%d rollbacks=%d", st.commits, st.rollbacks)
	}
}

func TestWithTx_RollbacksOnFnError(t *testing.T) {
	t.Parallel()
	st := &fakeState{}
	db := openFakeDB("tx_rollback", st)
	defer db.Close()

	fnErr := errors.New("boom")
	err := WithTx(context.Background(), db, TxOptions{}, func(ctx context.Context, tx *sql.Tx) error {
		return fnErr
	})
	if !errors.Is(err, fnErr) {
		t.Fatalf("expected fn error, got %v", err)
	}

	st.mu.Lock()
	defer st.mu.Unlock()
	if st.rollbacks != 1 {
		t.Fatalf("expected rollback, got rollbacks=%d", st.rollbacks)
	}
}

func TestWithTx_RollbackFailureWraps(t *testing.T) {
	t.Parallel()
	st := &fakeState{rollbackErr: errors.New("rbfail")}
	db := openFakeDB("tx_rb_fail", st)
	defer db.Close()

	fnErr := errors.New("boom")
	err := WithTx(context.Background(), db, TxOptions{}, func(ctx context.Context, tx *sql.Tx) error {
		return fnErr
	})
	if err == nil || !strings.Contains(err.Error(), "rollback failed") {
		t.Fatalf("expected wrapped rollback error, got %v", err)
	}
}

func TestWithTx_PanicRollbacksAndRethrows(t *testing.T) {
	t.Parallel()
	st := &fakeState{}
	db := openFakeDB("tx_panic", st)
	defer db.Close()

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
		st.mu.Lock()
		defer st.mu.Unlock()
		if st.rollbacks != 1 {
			t.Fatalf("expected rollback on panic, got %d", st.rollbacks)
		}
	}()

	_ = WithTx(context.Background(), db, TxOptions{}, func(ctx context.Context, tx *sql.Tx) error {
		panic("panic!")
	})
}
