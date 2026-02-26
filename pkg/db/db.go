package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Config struct {
	DSN string

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration

	PingTimeout time.Duration
}

func Open(cfg Config) (*sql.DB, error) {
	return openWithDriver("pgx", cfg)
}

func openWithDriver(driverName string, cfg Config) (*sql.DB, error) {
	if cfg.DSN == "" {
		return nil, errors.New("db: missing DSN")
	}
	if driverName == "" {
		return nil, errors.New("db: missing driver")
	}

	db, err := sql.Open(driverName, cfg.DSN)
	if err != nil {
		return nil, err
	}

	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}

	timeout := cfg.PingTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
