package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const schemaVersionTable = "schema_version"

func EnsureSchemaVersionTable(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return errors.New("db: nil db")
	}

	const ddl = `
CREATE TABLE IF NOT EXISTS schema_version (
	id INT PRIMARY KEY DEFAULT 1,
	version BIGINT NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO schema_version (id, version)
VALUES (1, 0)
ON CONFLICT (id) DO NOTHING;
`
	_, err := db.ExecContext(ctx, ddl)
	return err
}

func CurrentSchemaVersion(ctx context.Context, db *sql.DB) (int64, error) {
	if db == nil {
		return 0, errors.New("db: nil db")
	}

	var version int64
	err := db.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT version FROM %s WHERE id = 1`, schemaVersionTable),
	).Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

func EnforceForwardOnly(ctx context.Context, db *sql.DB, newVersion int64) error {
	if newVersion < 0 {
		return errors.New("db: invalid migration version")
	}

	cur, err := CurrentSchemaVersion(ctx, db)
	if err != nil {
		return err
	}

	if newVersion < cur {
		return fmt.Errorf("forward-only violation: current=%d new=%d", cur, newVersion)
	}

	return nil
}

func UpdateSchemaVersion(ctx context.Context, db *sql.DB, newVersion int64) error {
	if db == nil {
		return errors.New("db: nil db")
	}

	if err := EnforceForwardOnly(ctx, db, newVersion); err != nil {
		return err
	}

	const q = `
UPDATE schema_version
SET version = $1, updated_at = $2
WHERE id = 1;
`
	_, err := db.ExecContext(ctx, q, newVersion, time.Now().UTC())
	return err
}
