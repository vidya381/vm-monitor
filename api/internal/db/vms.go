package db

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vidya381/vm-monitor/api/internal/model"
)

type VMStore struct {
	pool *pgxpool.Pool
}

func NewVMStore(pool *pgxpool.Pool) *VMStore {
	return &VMStore{pool: pool}
}

func (s *VMStore) GetAll(ctx context.Context) ([]model.VM, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, address, auth_token, labels, status, last_heartbeat, created_at
		FROM vms ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vms []model.VM
	for rows.Next() {
		vm, err := scanVM(rows)
		if err != nil {
			return nil, err
		}
		vms = append(vms, vm)
	}
	return vms, rows.Err()
}

func (s *VMStore) GetByID(ctx context.Context, id uuid.UUID) (*model.VM, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, address, auth_token, labels, status, last_heartbeat, created_at
		FROM vms WHERE id = $1
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
	vm, err := scanVM(rows)
	if err != nil {
		return nil, err
	}
	return &vm, nil
}

// Upsert inserts or updates a VM by name. Returns the VM with its DB-assigned ID.
func (s *VMStore) Upsert(ctx context.Context, v model.VM) (*model.VM, error) {
	labelsJSON, err := json.Marshal(v.Labels)
	if err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		INSERT INTO vms (name, address, auth_token, labels)
		VALUES ($1, $2, $3, $4::jsonb)
		ON CONFLICT (name) DO UPDATE
			SET address    = EXCLUDED.address,
			    auth_token = EXCLUDED.auth_token,
			    labels     = EXCLUDED.labels
		RETURNING id, name, address, auth_token, labels, status, last_heartbeat, created_at
	`, v.Name, v.Address, v.AuthToken, labelsJSON)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("upsert returned no rows")
	}
	vm, err := scanVM(rows)
	if err != nil {
		return nil, err
	}
	return &vm, nil
}

// UpdateStatus sets the vm's status and last_heartbeat.
func (s *VMStore) UpdateStatus(ctx context.Context, id uuid.UUID, status string, heartbeat time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE vms SET status = $1, last_heartbeat = $2 WHERE id = $3
	`, status, heartbeat, pgtype.UUID{Bytes: id, Valid: true})
	return err
}

func scanVM(rows pgx.Rows) (model.VM, error) {
	var vm model.VM
	var pgID pgtype.UUID
	var labelsJSON []byte
	var lastHeartbeat pgtype.Timestamptz

	err := rows.Scan(
		&pgID, &vm.Name, &vm.Address, &vm.AuthToken,
		&labelsJSON, &vm.Status, &lastHeartbeat, &vm.CreatedAt,
	)
	if err != nil {
		return model.VM{}, err
	}

	vm.ID = uuid.UUID(pgID.Bytes)
	json.Unmarshal(labelsJSON, &vm.Labels)
	if vm.Labels == nil {
		vm.Labels = []string{}
	}
	if lastHeartbeat.Valid {
		t := lastHeartbeat.Time
		vm.LastHeartbeat = &t
	}
	return vm, nil
}
