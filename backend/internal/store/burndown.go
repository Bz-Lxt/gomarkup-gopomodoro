package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gopomodoro/internal/model"
	"gopomodoro/internal/timeutil"
)

func (d *DB) InsertBurndownPoint(ctx context.Context, p *model.BurndownPoint) error {
	_, err := d.SQL.ExecContext(ctx, `
		INSERT INTO burndown_points (id, milestone_id, recorded_at, remaining_points, ideal_points, event_type, event_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (event_id) DO NOTHING`,
		p.ID, p.MilestoneID, p.RecordedAt, p.RemainingPoints, p.IdealPoints, p.EventType, p.EventID)
	return err
}

func (d *DB) ListBurndownPoints(ctx context.Context, milestoneID model.ID, from, to time.Time) ([]model.BurndownPoint, error) {
	rows, err := d.SQL.QueryContext(ctx, `
		SELECT id, milestone_id, recorded_at, remaining_points, ideal_points, event_type, event_id
		FROM burndown_points
		WHERE milestone_id=$1 AND recorded_at >= $2 AND recorded_at <= $3
		ORDER BY recorded_at ASC`, milestoneID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.BurndownPoint
	for rows.Next() {
		var p model.BurndownPoint
		if err := rows.Scan(&p.ID, &p.MilestoneID, &p.RecordedAt, &p.RemainingPoints, &p.IdealPoints, &p.EventType, &p.EventID); err != nil {
			return nil, err
		}
		p.RecordedAt = timeutil.ToBeijing(p.RecordedAt)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (d *DB) InsertScopeChange(ctx context.Context, l *model.ScopeChangeLog) error {
	_, err := d.SQL.ExecContext(ctx, `
		INSERT INTO scope_change_logs (id, milestone_id, delta_points, reason, occurred_at)
		VALUES ($1,$2,$3,$4,$5)`,
		l.ID, l.MilestoneID, l.DeltaPoints, l.Reason, l.OccurredAt)
	return err
}

func (d *DB) TryClaimEvent(ctx context.Context, eventID, eventType string, at time.Time) (bool, error) {
	res, err := d.SQL.ExecContext(ctx, `
		INSERT INTO processed_events (event_id, event_type, processed_at)
		VALUES ($1,$2,$3)
		ON CONFLICT (event_id) DO NOTHING`, eventID, eventType, at)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n == 1, nil
}

func (d *DB) EventProcessed(ctx context.Context, eventID string) (bool, error) {
	var n int
	err := d.SQL.QueryRowContext(ctx, `SELECT COUNT(1) FROM processed_events WHERE event_id=$1`, eventID).Scan(&n)
	return n > 0, err
}

func NullID(id *model.ID) any {
	if id == nil {
		return nil
	}
	return *id
}

func IsNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
