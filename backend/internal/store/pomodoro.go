package store

import (
	"context"
	"database/sql"
	"errors"

	"gopomodoro/internal/httpx"
	"gopomodoro/internal/model"
	"gopomodoro/internal/timeutil"
)

func (d *DB) InsertSession(ctx context.Context, s *model.PomodoroSession) error {
	_, err := d.SQL.ExecContext(ctx, `
		INSERT INTO pomodoro_sessions
		(id, user_id, task_id, state, focus_duration_ms, started_at, paused_at, paused_accumulated_ms,
		 expected_end_at, ended_at, abort_reason, resume_token, version, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		s.ID, s.UserID, s.TaskID, s.State, s.FocusDurationMS, s.StartedAt, s.PausedAt, s.PausedAccumulatedMS,
		s.ExpectedEndAt, s.EndedAt, s.AbortReason, s.ResumeToken, s.Version, s.CreatedAt, s.UpdatedAt)
	return err
}

func (d *DB) UpdateSessionOptimistic(ctx context.Context, s *model.PomodoroSession, prevVersion int64) error {
	res, err := d.SQL.ExecContext(ctx, `
		UPDATE pomodoro_sessions SET
			state=$1, started_at=$2, paused_at=$3, paused_accumulated_ms=$4,
			expected_end_at=$5, ended_at=$6, abort_reason=$7, version=$8, updated_at=$9
		WHERE id=$10 AND version=$11`,
		s.State, s.StartedAt, s.PausedAt, s.PausedAccumulatedMS,
		s.ExpectedEndAt, s.EndedAt, s.AbortReason, s.Version, s.UpdatedAt,
		s.ID, prevVersion)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return httpx.ErrOptimisticLock
	}
	return nil
}

func (d *DB) SessionByID(ctx context.Context, id model.ID) (*model.PomodoroSession, error) {
	return scanSession(d.SQL.QueryRowContext(ctx, sessionSelect+` WHERE id=$1`, id))
}

func (d *DB) SessionByResumeToken(ctx context.Context, token string) (*model.PomodoroSession, error) {
	return scanSession(d.SQL.QueryRowContext(ctx, sessionSelect+` WHERE resume_token=$1`, token))
}

func (d *DB) ActiveSessionByUser(ctx context.Context, userID model.ID) (*model.PomodoroSession, error) {
	s, err := scanSession(d.SQL.QueryRowContext(ctx, sessionSelect+`
		WHERE user_id=$1 AND state IN ('running','paused') ORDER BY created_at DESC LIMIT 1`, userID))
	if err != nil {
		if errors.Is(err, httpx.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return s, nil
}

func (d *DB) ListLiveSessions(ctx context.Context) ([]model.PomodoroSession, error) {
	rows, err := d.SQL.QueryContext(ctx, sessionSelect+` WHERE state IN ('running','paused')`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.PomodoroSession
	for rows.Next() {
		s, err := scanSessionRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *s)
	}
	return out, rows.Err()
}

func (d *DB) CountSessions(ctx context.Context, userID model.ID, states []string, from, to any) (int, error) {
	q := `SELECT COUNT(1) FROM pomodoro_sessions WHERE user_id=$1 AND state = ANY($2) AND ended_at >= $3 AND ended_at < $4`
	var n int
	err := d.SQL.QueryRowContext(ctx, q, userID, states, from, to).Scan(&n)
	return n, err
}

func (d *DB) CountSessionsSince(ctx context.Context, userID model.ID, state string, from any) (int, error) {
	var n int
	err := d.SQL.QueryRowContext(ctx, `
		SELECT COUNT(1) FROM pomodoro_sessions
		WHERE user_id=$1 AND state=$2 AND ended_at >= $3`, userID, state, from).Scan(&n)
	return n, err
}

const sessionSelect = `
	SELECT id, user_id, task_id, state, focus_duration_ms, started_at, paused_at, paused_accumulated_ms,
	       expected_end_at, ended_at, abort_reason, resume_token, version, created_at, updated_at
	FROM pomodoro_sessions`

func scanSession(row *sql.Row) (*model.PomodoroSession, error) {
	s, err := scanSessionDest(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, httpx.ErrNotFound
	}
	return s, err
}

func scanSessionRow(rows *sql.Rows) (*model.PomodoroSession, error) {
	return scanSessionDest(rows)
}

func scanSessionDest(s scanner) (*model.PomodoroSession, error) {
	var p model.PomodoroSession
	if err := s.Scan(&p.ID, &p.UserID, &p.TaskID, &p.State, &p.FocusDurationMS, &p.StartedAt, &p.PausedAt,
		&p.PausedAccumulatedMS, &p.ExpectedEndAt, &p.EndedAt, &p.AbortReason, &p.ResumeToken,
		&p.Version, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	p.CreatedAt = timeutil.ToBeijing(p.CreatedAt)
	p.UpdatedAt = timeutil.ToBeijing(p.UpdatedAt)
	if p.StartedAt != nil {
		t := timeutil.ToBeijing(*p.StartedAt)
		p.StartedAt = &t
	}
	if p.PausedAt != nil {
		t := timeutil.ToBeijing(*p.PausedAt)
		p.PausedAt = &t
	}
	if p.ExpectedEndAt != nil {
		t := timeutil.ToBeijing(*p.ExpectedEndAt)
		p.ExpectedEndAt = &t
	}
	if p.EndedAt != nil {
		t := timeutil.ToBeijing(*p.EndedAt)
		p.EndedAt = &t
	}
	return &p, nil
}
