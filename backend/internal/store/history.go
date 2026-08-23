package store

import (
	"context"
	"time"

	"gopomodoro/internal/model"
	"gopomodoro/internal/timeutil"
)

type SessionFilter struct {
	UserID model.ID
	State  string
	From   time.Time
	To     time.Time
	Limit  int
}

func (d *DB) ListSessions(ctx context.Context, f SessionFilter) ([]model.PomodoroSession, error) {
	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 50
	}
	q := sessionSelect + ` WHERE user_id=$1`
	args := []any{f.UserID}
	n := 2
	if f.State != "" {
		q += ` AND state=$` + itoa(n)
		args = append(args, f.State)
		n++
	}
	if !f.From.IsZero() {
		q += ` AND created_at >= $` + itoa(n)
		args = append(args, f.From)
		n++
	}
	if !f.To.IsZero() {
		q += ` AND created_at < $` + itoa(n)
		args = append(args, f.To)
		n++
	}
	q += ` ORDER BY created_at DESC LIMIT $` + itoa(n)
	args = append(args, f.Limit)
	rows, err := d.SQL.QueryContext(ctx, q, args...)
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

func (d *DB) DailyCompleted(ctx context.Context, userID model.ID, from, to time.Time) ([]struct {
	Day   time.Time
	Count int
}, error) {
	rows, err := d.SQL.QueryContext(ctx, `
		SELECT date_trunc('day', ended_at AT TIME ZONE 'Asia/Shanghai') AS d, COUNT(1)
		FROM pomodoro_sessions
		WHERE user_id=$1 AND state='completed' AND ended_at >= $2 AND ended_at < $3
		GROUP BY 1 ORDER BY 1`, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		Day   time.Time
		Count int
	}
	for rows.Next() {
		var day time.Time
		var n int
		if err := rows.Scan(&day, &n); err != nil {
			return nil, err
		}
		out = append(out, struct {
			Day   time.Time
			Count int
		}{Day: timeutil.StartOfDay(day), Count: n})
	}
	return out, rows.Err()
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return string([]byte{'0' + byte(n/10), '0' + byte(n%10)})
}
