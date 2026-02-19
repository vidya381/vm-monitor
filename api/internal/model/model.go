package model

import (
	"time"

	"github.com/google/uuid"
)

type VM struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	Address       string     `json:"address"`
	AuthToken     string     `json:"-"` // never exposed in API responses
	Labels        []string   `json:"labels"`
	Status        string     `json:"status"`
	LastHeartbeat *time.Time `json:"last_heartbeat"`
	CreatedAt     time.Time  `json:"created_at"`
}

type App struct {
	ID              uuid.UUID  `json:"id"`
	VMID            uuid.UUID  `json:"vm_id"`
	VMName          string     `json:"vm_name"`
	Name            string     `json:"name"`
	Type            string     `json:"type"`
	Environment     string     `json:"environment,omitempty"`
	Config          AppConfig  `json:"config"`
	LastStatus      string     `json:"last_status"`
	LastCheckedAt   *time.Time `json:"last_checked_at"`
	LastRestartedAt *time.Time `json:"last_restarted_at"`
	CreatedAt       time.Time  `json:"created_at"`

	// Internal — used for proxying, never in JSON
	VMAddress   string `json:"-"`
	VMAuthToken string `json:"-"`
}

type AppConfig struct {
	Service     string      `json:"service"`
	EnvFile     string      `json:"env_file,omitempty"`
	HealthCheck HealthCheck `json:"health_check,omitempty"`
}

type HealthCheck struct {
	Type string `json:"type,omitempty"`
	URL  string `json:"url,omitempty"`
	Cmd  string `json:"cmd,omitempty"`
}

type AuditLog struct {
	ID        int64          `json:"id"`
	AppID     *uuid.UUID     `json:"app_id"`
	Action    string         `json:"action"`
	Details   map[string]any `json:"details,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

// RegisterRequest is the payload the agent sends on startup.
type RegisterRequest struct {
	Name      string      `json:"name"`
	Address   string      `json:"address"`
	AuthToken string      `json:"auth_token"`
	Labels    []string    `json:"labels"`
	Apps      []AppInput  `json:"apps"`
}

type AppInput struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Environment string      `json:"environment"`
	Config      AppConfig   `json:"config"`
}
