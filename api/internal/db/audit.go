package db

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vidya381/vm-monitor/api/internal/model"
)

type AuditStore struct {
	pool *pgxpool.Pool
}

func NewAuditStore(pool *pgxpool.Pool) *AuditStore {
	return &AuditStore{pool: pool}
}

func (s *AuditStore) Create(ctx context.Context, appID uuid.UUID, action string, details map[string]any) error {
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO audit_logs (app_id, action, details)
		VALUES ($1, $2, $3::jsonb)
	`, pgtype.UUID{Bytes: appID, Valid: true}, action, detailsJSON)
	return err
}

func (s *AuditStore) GetByAppID(ctx context.Context, appID uuid.UUID) ([]model.AuditLog, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, app_id, action, details, created_at
		FROM audit_logs WHERE app_id = $1
		ORDER BY created_at DESC LIMIT 100
	`, pgtype.UUID{Bytes: appID, Valid: true})
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAuditLogs(rows)
}

func (s *AuditStore) GetAll(ctx context.Context) ([]model.AuditLog, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, app_id, action, details, created_at
		FROM audit_logs ORDER BY created_at DESC LIMIT 200
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAuditLogs(rows)
}

func scanAuditLogs(rows pgx.Rows) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	for rows.Next() {
		var l model.AuditLog
		var pgAppID pgtype.UUID
		var detailsJSON []byte

		if err := rows.Scan(&l.ID, &pgAppID, &l.Action, &detailsJSON, &l.CreatedAt); err != nil {
			return nil, err
		}
		if pgAppID.Valid {
			id := uuid.UUID(pgAppID.Bytes)
			l.AppID = &id
		}
		json.Unmarshal(detailsJSON, &l.Details)
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
