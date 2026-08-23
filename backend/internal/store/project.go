package store

import (
	"context"
	"database/sql"
	"errors"

	"gopomodoro/internal/httpx"
	"gopomodoro/internal/model"
	"gopomodoro/internal/timeutil"
)

func (d *DB) CreateProject(ctx context.Context, p *model.Project) error {
	_, err := d.SQL.ExecContext(ctx, `
		INSERT INTO projects (id, user_id, name, description, archived, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		p.ID, p.UserID, p.Name, p.Description, p.Archived, p.CreatedAt, p.UpdatedAt)
	return err
}

func (d *DB) UpdateProject(ctx context.Context, p *model.Project) error {
	res, err := d.SQL.ExecContext(ctx, `
		UPDATE projects SET name=$1, description=$2, archived=$3, updated_at=$4
		WHERE id=$5 AND user_id=$6`,
		p.Name, p.Description, p.Archived, p.UpdatedAt, p.ID, p.UserID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return httpx.ErrNotFound
	}
	return nil
}

func (d *DB) ProjectByID(ctx context.Context, userID, id model.ID) (*model.Project, error) {
	return scanProject(d.SQL.QueryRowContext(ctx, `
		SELECT id, user_id, name, description, archived, created_at, updated_at
		FROM projects WHERE id=$1 AND user_id=$2`, id, userID))
}

func (d *DB) ListProjects(ctx context.Context, userID model.ID, includeArchived bool) ([]model.Project, error) {
	q := `SELECT id, user_id, name, description, archived, created_at, updated_at
		FROM projects WHERE user_id=$1`
	if !includeArchived {
		q += ` AND archived=false`
	}
	q += ` ORDER BY created_at DESC`
	rows, err := d.SQL.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Project
	for rows.Next() {
		p, err := scanProjectRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

func scanProject(row *sql.Row) (*model.Project, error) {
	var p model.Project
	err := row.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.Archived, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, httpx.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.CreatedAt = timeutil.ToBeijing(p.CreatedAt)
	p.UpdatedAt = timeutil.ToBeijing(p.UpdatedAt)
	return &p, nil
}

func scanProjectRow(rows *sql.Rows) (*model.Project, error) {
	var p model.Project
	if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.Archived, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	p.CreatedAt = timeutil.ToBeijing(p.CreatedAt)
	p.UpdatedAt = timeutil.ToBeijing(p.UpdatedAt)
	return &p, nil
}
