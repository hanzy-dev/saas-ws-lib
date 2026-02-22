package db

import (
	"context"
	"database/sql"
	"fmt"
)

type TxFunc func(ctx context.Context, tx *sql.Tx) error

type TxOptions struct {
	Isolation sql.IsolationLevel
	ReadOnly  bool
}

func WithTx(ctx context.Context, db *sql.DB, opts TxOptions, fn TxFunc) error {
	if db == nil {
		return sql.ErrConnDone
	}
	if fn == nil {
		return fmt.Errorf("db: nil tx function")
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		Isolation: opts.Isolation,
		ReadOnly:  opts.ReadOnly,
	})
	if err != nil {
		return fmt.Errorf("db: begin tx: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(ctx, tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("db: rollback failed: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("db: commit failed: %w", err)
	}

	return nil
}
