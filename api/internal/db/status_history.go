package db

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vidya381/vm-monitor/api/internal/model"
)

type StatusHistoryStore struct {
	pool *pgxpool.Pool
}

func NewStatusHistoryStore(pool *pgxpool.Pool) *StatusHistoryStore {
	return &StatusHistoryStore{pool: pool}
}

// RecordTransition closes the current open row and inserts a new one for newStatus.
func (s *StatusHistoryStore) RecordTransition(ctx context.Context, appID uuid.UUID, newStatus string) error {
	pgID := pgtype.UUID{Bytes: appID, Valid: true}
	_, err := s.pool.Exec(ctx, `
		UPDATE status_history
		SET ended_at   = NOW(),
		    duration_s = EXTRACT(EPOCH FROM (NOW() - started_at))::int
		WHERE app_id = $1 AND ended_at IS NULL
	`, pgID)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO status_history (app_id, status, started_at)
		VALUES ($1, $2, NOW())
	`, pgID, newStatus)
	return err
}

// GetUptimeResponse returns uptime statistics and incident list for the given window.
// UptimePct is nil if no history exists yet.
func (s *StatusHistoryStore) GetUptimeResponse(ctx context.Context, appID uuid.UUID, windowStart time.Time, days int) (model.UptimeResponse, error) {
	pgID := pgtype.UUID{Bytes: appID, Valid: true}

	resp := model.UptimeResponse{
		AppID:      appID.String(),
		WindowDays: days,
		Incidents:  []model.Incident{},
	}

	// Check if any rows exist in the window.
	var count int
	if err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM status_history WHERE app_id = $1 AND started_at >= $2`,
		pgID, windowStart,
	).Scan(&count); err != nil {
		return resp, err
	}
	if count == 0 {
		return resp, nil
	}

	// Calculate uptime percentage.
	var runningSeconds, windowSeconds float64
	if err := s.pool.QueryRow(ctx, `
		SELECT
		  COALESCE(SUM(
		    EXTRACT(EPOCH FROM (COALESCE(ended_at, NOW()) - started_at))
		  ) FILTER (WHERE status = 'running'), 0),
		  EXTRACT(EPOCH FROM (NOW() - $2::timestamptz))
		FROM status_history
		WHERE app_id = $1 AND started_at >= $2
	`, pgID, windowStart).Scan(&runningSeconds, &windowSeconds); err != nil {
		return resp, err
	}

	pct := 100.0
	if windowSeconds > 0 {
		pct = math.Round((runningSeconds/windowSeconds)*1000) / 10
		if pct > 100 {
			pct = 100
		}
	}
	resp.UptimePct = &pct

	// Fetch non-running incidents.
	rows, err := s.pool.Query(ctx, `
		SELECT status, started_at,
		       COALESCE(ended_at, NOW()),
		       COALESCE(duration_s, EXTRACT(EPOCH FROM (NOW() - started_at))::int)
		FROM status_history
		WHERE app_id = $1
		  AND status != 'running'
		  AND started_at >= $2
		ORDER BY started_at DESC
	`, pgID, windowStart)
	if err != nil {
		return resp, err
	}
	defer rows.Close()

	var totalDowntimeS int
	for rows.Next() {
		var inc model.Incident
		var startedAt, endedAt pgtype.Timestamptz
		var durationS int
		if err := rows.Scan(&inc.Status, &startedAt, &endedAt, &durationS); err != nil {
			return resp, err
		}
		inc.StartedAt = startedAt.Time
		inc.EndedAt = endedAt.Time
		inc.DurationS = durationS
		resp.Incidents = append(resp.Incidents, inc)
		totalDowntimeS += durationS
	}
	if err := rows.Err(); err != nil {
		return resp, err
	}

	resp.IncidentCount = len(resp.Incidents)
	resp.TotalDowntimeS = totalDowntimeS
	return resp, nil
}

// Purge deletes rows older than 90 days.
func (s *StatusHistoryStore) Purge(ctx context.Context) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM status_history WHERE started_at < NOW() - INTERVAL '90 days'`)
	return err
}
