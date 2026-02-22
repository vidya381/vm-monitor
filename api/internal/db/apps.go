package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vidya381/vm-monitor/api/internal/model"
)

type AppStore struct {
	pool *pgxpool.Pool
}

func NewAppStore(pool *pgxpool.Pool) *AppStore {
	return &AppStore{pool: pool}
}

// GetAll returns all apps joined with their VM (for address/token proxying).
func (s *AppStore) GetAll(ctx context.Context) ([]model.App, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT a.id, a.vm_id, v.name, v.address, v.auth_token,
		       a.name, a.type, a.environment, a.config,
		       a.last_status, a.last_checked_at, a.last_restarted_at, a.created_at
		FROM apps a
		JOIN vms v ON v.id = a.vm_id
		ORDER BY v.name, a.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []model.App
	for rows.Next() {
		app, err := scanApp(rows)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}
	return apps, rows.Err()
}

func (s *AppStore) GetByID(ctx context.Context, id uuid.UUID) (*model.App, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT a.id, a.vm_id, v.name, v.address, v.auth_token,
		       a.name, a.type, a.environment, a.config,
		       a.last_status, a.last_checked_at, a.last_restarted_at, a.created_at
		FROM apps a
		JOIN vms v ON v.id = a.vm_id
		WHERE a.id = $1
	`, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, ErrNotFound
	}
	app, err := scanApp(rows)
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// Upsert inserts or updates an app by (vm_id, name).
func (s *AppStore) Upsert(ctx context.Context, vmID uuid.UUID, input model.AppInput) error {
	configJSON, err := json.Marshal(input.Config)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO apps (vm_id, name, type, environment, config)
		VALUES ($1, $2, $3, $4, $5::jsonb)
		ON CONFLICT (vm_id, name) DO UPDATE
			SET type        = EXCLUDED.type,
			    environment = EXCLUDED.environment,
			    config      = EXCLUDED.config
	`,
		pgtype.UUID{Bytes: vmID, Valid: true},
		input.Name, input.Type, input.Environment, configJSON,
	)
	return err
}

// GetByVMAndName returns an app by its VM ID and name.
func (s *AppStore) GetByVMAndName(ctx context.Context, vmID uuid.UUID, name string) (*model.App, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT a.id, a.vm_id, v.name, v.address, v.auth_token,
		       a.name, a.type, a.environment, a.config,
		       a.last_status, a.last_checked_at, a.last_restarted_at, a.created_at
		FROM apps a
		JOIN vms v ON v.id = a.vm_id
		WHERE a.vm_id = $1 AND a.name = $2
	`, pgtype.UUID{Bytes: vmID, Valid: true}, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, ErrNotFound
	}
	app, err := scanApp(rows)
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// UpdateStatus sets last_status and last_checked_at for an app found by (vm_id, app name).
func (s *AppStore) UpdateStatus(ctx context.Context, vmID uuid.UUID, appName, status string, checkedAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE apps SET last_status = $1, last_checked_at = $2
		WHERE vm_id = $3 AND name = $4
	`, status, checkedAt, pgtype.UUID{Bytes: vmID, Valid: true}, appName)
	return err
}

// UpdateLastRestarted sets last_restarted_at for an app.
func (s *AppStore) UpdateLastRestarted(ctx context.Context, id uuid.UUID, t time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE apps SET last_restarted_at = $1 WHERE id = $2
	`, t, pgtype.UUID{Bytes: id, Valid: true})
	return err
}

func scanApp(rows pgx.Rows) (model.App, error) {
	var app model.App
	var pgID, pgVMID pgtype.UUID
	var configJSON []byte
	var lastChecked, lastRestarted pgtype.Timestamptz

	err := rows.Scan(
		&pgID, &pgVMID, &app.VMName, &app.VMAddress, &app.VMAuthToken,
		&app.Name, &app.Type, &app.Environment, &configJSON,
		&app.LastStatus, &lastChecked, &lastRestarted, &app.CreatedAt,
	)
	if err != nil {
		return model.App{}, err
	}

	app.ID = uuid.UUID(pgID.Bytes)
	app.VMID = uuid.UUID(pgVMID.Bytes)
	json.Unmarshal(configJSON, &app.Config)

	if lastChecked.Valid {
		t := lastChecked.Time
		app.LastCheckedAt = &t
	}
	if lastRestarted.Valid {
		t := lastRestarted.Time
		app.LastRestartedAt = &t
	}
	return app, nil
}

