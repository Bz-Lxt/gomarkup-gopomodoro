package store

import (
	"context"
	"database/sql"
	"errors"

	"gopomodoro/internal/httpx"
	"gopomodoro/internal/model"
	"gopomodoro/internal/timeutil"
)

func (d *DB) CreateUser(ctx context.Context, u *model.User) error {
	_, err := d.SQL.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, display_name, timezone, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		u.ID, u.Email, u.PasswordHash, u.DisplayName, u.Timezone, u.CreatedAt)
	if err != nil {
		if isUnique(err) {
			return httpx.ErrConflict.WithDetails(map[string]any{"field": "email"})
		}
		return err
	}
	return nil
}

func (d *DB) UserByEmail(ctx context.Context, email string) (*model.User, error) {
	return scanUser(d.SQL.QueryRowContext(ctx, `
		SELECT id, email, password_hash, display_name, timezone, created_at
		FROM users WHERE email=$1`, email))
}

func (d *DB) UserByID(ctx context.Context, id model.ID) (*model.User, error) {
	return scanUser(d.SQL.QueryRowContext(ctx, `
		SELECT id, email, password_hash, display_name, timezone, created_at
		FROM users WHERE id=$1`, id))
}

func (d *DB) CountUsers(ctx context.Context) (int, error) {
	var n int
	err := d.SQL.QueryRowContext(ctx, `SELECT COUNT(1) FROM users`).Scan(&n)
	return n, err
}

func scanUser(row *sql.Row) (*model.User, error) {
	var u model.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Timezone, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, httpx.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	u.CreatedAt = timeutil.ToBeijing(u.CreatedAt)
	return &u, nil
}
