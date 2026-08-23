package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"gopomodoro/internal/logger"
)

type DB struct {
	SQL *sql.DB
}

func Open(ctx context.Context, dsn string) (*DB, error) {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	logger.L().Info("database connected")
	return &DB{SQL: sqlDB}, nil
}

func (d *DB) Close() error { return d.SQL.Close() }

func (d *DB) Tx(ctx context.Context, fn func(*sql.Tx) error) error {
	// Reject already-canceled contexts before acquiring a connection
	// or beginning a transaction. A timed-out / disconnected request
	// must never start a transaction, let alone run the callback or commit.
	if err := ctx.Err(); err != nil {
		return err
	}
	tx, err := d.SQL.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	// Re-check after the callback succeeds: if the request was
	// canceled while the callback ran, roll back instead of committing.
	// tx.Commit() does not accept a context, so we must guard explicitly.
	if err := ctx.Err(); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
