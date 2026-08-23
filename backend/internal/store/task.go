package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"gopomodoro/internal/httpx"
	"gopomodoro/internal/model"
	"gopomodoro/internal/timeutil"
)

func (d *DB) CreateTask(ctx context.Context, t *model.Task) error {
	_, err := d.SQL.ExecContext(ctx, `
		INSERT INTO tasks (id, project_id, milestone_id, title, estimated_pomodoros, consumed_pomodoros, kanban_column, sort_order, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		t.ID, t.ProjectID, t.MilestoneID, t.Title, t.EstimatedPomodoros, t.ConsumedPomodoros, t.KanbanColumn, t.SortOrder, t.CreatedAt, t.UpdatedAt)
	return err
}

func (d *DB) UpdateTask(ctx context.Context, t *model.Task) error {
	_, err := d.SQL.ExecContext(ctx, `
		UPDATE tasks SET title=$1, estimated_pomodoros=$2, consumed_pomodoros=$3, kanban_column=$4,
			sort_order=$5, milestone_id=$6, updated_at=$7
		WHERE id=$8`,
		t.Title, t.EstimatedPomodoros, t.ConsumedPomodoros, t.KanbanColumn, t.SortOrder, t.MilestoneID, t.UpdatedAt, t.ID)
	return err
}

func (d *DB) DeleteTask(ctx context.Context, id model.ID) error {
	res, err := d.SQL.ExecContext(ctx, `DELETE FROM tasks WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return httpx.ErrNotFound
	}
	return nil
}

func (d *DB) TaskByID(ctx context.Context, id model.ID) (*model.Task, error) {
	return scanTask(d.SQL.QueryRowContext(ctx, `
		SELECT id, project_id, milestone_id, title, estimated_pomodoros, consumed_pomodoros, kanban_column, sort_order, created_at, updated_at
		FROM tasks WHERE id=$1`, id))
}

func (d *DB) TaskOwnedBy(ctx context.Context, userID, taskID model.ID) (*model.Task, error) {
	return scanTask(d.SQL.QueryRowContext(ctx, `
		SELECT t.id, t.project_id, t.milestone_id, t.title, t.estimated_pomodoros, t.consumed_pomodoros, t.kanban_column, t.sort_order, t.created_at, t.updated_at
		FROM tasks t JOIN projects p ON p.id=t.project_id
		WHERE t.id=$1 AND p.user_id=$2`, taskID, userID))
}

func (d *DB) ListTasks(ctx context.Context, projectID model.ID, milestoneID *model.ID) ([]model.Task, error) {
	q := `
		SELECT id, project_id, milestone_id, title, estimated_pomodoros, consumed_pomodoros, kanban_column, sort_order, created_at, updated_at
		FROM tasks WHERE project_id=$1`
	args := []any{projectID}
	if milestoneID != nil {
		q += ` AND milestone_id=$2`
		args = append(args, *milestoneID)
	}
	q += ` ORDER BY sort_order ASC, created_at ASC`
	rows, err := d.SQL.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Task
	for rows.Next() {
		t, err := scanTaskRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

func (d *DB) ListTasksByMilestone(ctx context.Context, milestoneID model.ID) ([]model.Task, error) {
	rows, err := d.SQL.QueryContext(ctx, `
		SELECT id, project_id, milestone_id, title, estimated_pomodoros, consumed_pomodoros, kanban_column, sort_order, created_at, updated_at
		FROM tasks WHERE milestone_id=$1`, milestoneID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Task
	for rows.Next() {
		t, err := scanTaskRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

func (d *DB) IncrementConsumed(ctx context.Context, taskID model.ID) error {
	_, err := d.SQL.ExecContext(ctx, `
		UPDATE tasks SET consumed_pomodoros = consumed_pomodoros + 1, updated_at = NOW()
		WHERE id=$1`, taskID)
	return err
}

func (d *DB) ReorderTasks(ctx context.Context, updates []struct {
	ID     uuid.UUID
	Column model.KanbanColumn
	Order  int
}) error {
	return d.Tx(ctx, func(tx *sql.Tx) error {
		for _, u := range updates {
			if _, err := tx.ExecContext(ctx, `
				UPDATE tasks SET kanban_column=$1, sort_order=$2, updated_at=NOW() WHERE id=$3`,
				u.Column, u.Order, u.ID); err != nil {
				return err
			}
		}
		return nil
	})
}

func scanTask(row *sql.Row) (*model.Task, error) {
	t, err := scanTaskDest(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, httpx.ErrNotFound
	}
	return t, err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTaskDest(s scanner) (*model.Task, error) {
	var t model.Task
	var mid sql.NullString
	if err := s.Scan(&t.ID, &t.ProjectID, &mid, &t.Title, &t.EstimatedPomodoros, &t.ConsumedPomodoros, &t.KanbanColumn, &t.SortOrder, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, err
	}
	if mid.Valid {
		id, err := uuid.Parse(mid.String)
		if err == nil {
			t.MilestoneID = &id
		}
	}
	t.CreatedAt = timeutil.ToBeijing(t.CreatedAt)
	t.UpdatedAt = timeutil.ToBeijing(t.UpdatedAt)
	return &t, nil
}

func scanTaskRow(rows *sql.Rows) (*model.Task, error) {
	return scanTaskDest(rows)
}
