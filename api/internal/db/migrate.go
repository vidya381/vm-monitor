package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations creates all tables if they don't already exist.
// Idempotent — safe to run on every startup.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS vms (
		  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		  name           VARCHAR(255) NOT NULL UNIQUE,
		  address        VARCHAR(255) NOT NULL,
		  auth_token     VARCHAR(512) NOT NULL,
		  labels         JSONB DEFAULT '[]',
		  status         VARCHAR(50) DEFAULT 'unknown',
		  last_heartbeat TIMESTAMPTZ,
		  created_at     TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS apps (
		  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		  vm_id             UUID REFERENCES vms(id) ON DELETE CASCADE,
		  name              VARCHAR(255) NOT NULL,
		  type              VARCHAR(50) NOT NULL,
		  environment       VARCHAR(50),
		  config            JSONB NOT NULL DEFAULT '{}',
		  last_status       VARCHAR(50),
		  last_checked_at   TIMESTAMPTZ,
		  last_restarted_at TIMESTAMPTZ,
		  created_at        TIMESTAMPTZ DEFAULT NOW(),
		  UNIQUE(vm_id, name)
		);

		CREATE TABLE IF NOT EXISTS audit_logs (
		  id         BIGSERIAL PRIMARY KEY,
		  app_id     UUID REFERENCES apps(id) ON DELETE SET NULL,
		  action     VARCHAR(100) NOT NULL,
		  details    JSONB,
		  created_at TIMESTAMPTZ DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}
	return nil
}
