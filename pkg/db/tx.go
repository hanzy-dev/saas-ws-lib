package db

import (
	"context"
	"database/sql"
)

type TxFunc func(ctx context.Context, tx *sql.Tx) error

func WithTx(ctx context.Context, db *sql.DB, fn TxFunc) error {
	if db == nil {
		return sql.ErrConnDone
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		// If panic occurs, rollback and re-panic
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(ctx, tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
