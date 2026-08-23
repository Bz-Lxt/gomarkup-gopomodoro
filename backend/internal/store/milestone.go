package store

import (
	"context"
	"database/sql"
	"errors"

	"gopomodoro/internal/httpx"
	"gopomodoro/internal/model"
	"gopomodoro/internal/timeutil"
)

func (d *DB) CreateMilestone(ctx context.Context, m *model.Milestone) error {
	_, err := d.SQL.ExecContext(ctx, `
		INSERT INTO milestones (id, project_id, title, start_date, due_date, baseline_points, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		m.ID, m.ProjectID, m.Title, m.StartDate, m.DueDate, m.BaselinePoints, m.Status, m.CreatedAt, m.UpdatedAt)
	return err
}

func (d *DB) UpdateMilestone(ctx context.Context, m *model.Milestone) error {
	_, err := d.SQL.ExecContext(ctx, `
		UPDATE milestones SET title=$1, start_date=$2, due_date=$3, baseline_points=$4, status=$5, updated_at=$6
		WHERE id=$7`,
		m.Title, m.StartDate, m.DueDate, m.BaselinePoints, m.Status, m.UpdatedAt, m.ID)
	return err
}

func (d *DB) MilestoneByID(ctx context.Context, id model.ID) (*model.Milestone, error) {
	return scanMilestone(d.SQL.QueryRowContext(ctx, `
		SELECT id, project_id, title, start_date, due_date, baseline_points, status, created_at, updated_at
		FROM milestones WHERE id=$1`, id))
}

func (d *DB) ListMilestones(ctx context.Context, projectID model.ID) ([]model.Milestone, error) {
	rows, err := d.SQL.QueryContext(ctx, `
		SELECT id, project_id, title, start_date, due_date, baseline_points, status, created_at, updated_at
		FROM milestones WHERE project_id=$1 ORDER BY due_date ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Milestone
	for rows.Next() {
		m, err := scanMilestoneRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	return out, rows.Err()
}

func (d *DB) MilestoneOwnedBy(ctx context.Context, userID, milestoneID model.ID) (*model.Milestone, error) {
	row := d.SQL.QueryRowContext(ctx, `
		SELECT m.id, m.project_id, m.title, m.start_date, m.due_date, m.baseline_points, m.status, m.created_at, m.updated_at
		FROM milestones m
		JOIN projects p ON p.id = m.project_id
		WHERE m.id=$1 AND p.user_id=$2`, milestoneID, userID)
	return scanMilestone(row)
}

func scanMilestone(row *sql.Row) (*model.Milestone, error) {
	var m model.Milestone
	err := row.Scan(&m.ID, &m.ProjectID, &m.Title, &m.StartDate, &m.DueDate, &m.BaselinePoints, &m.Status, &m.CreatedAt, &m.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, httpx.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	normalizeMilestone(&m)
	return &m, nil
}

func scanMilestoneRow(rows *sql.Rows) (*model.Milestone, error) {
	var m model.Milestone
	if err := rows.Scan(&m.ID, &m.ProjectID, &m.Title, &m.StartDate, &m.DueDate, &m.BaselinePoints, &m.Status, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return nil, err
	}
	normalizeMilestone(&m)
	return &m, nil
}

func normalizeMilestone(m *model.Milestone) {
	m.StartDate = timeutil.StartOfDay(m.StartDate)
	m.DueDate = timeutil.StartOfDay(m.DueDate)
	m.CreatedAt = timeutil.ToBeijing(m.CreatedAt)
	m.UpdatedAt = timeutil.ToBeijing(m.UpdatedAt)
}
