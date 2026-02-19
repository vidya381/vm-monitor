package config

import (
	"os"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "agent*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_Valid(t *testing.T) {
	path := writeTemp(t, `
vm:
  name: "test-vm"
  port: 9000
  control_plane_url: "http://localhost:8080"
  auth_token: "secret"
  labels:
    - "production"

apps:
  - name: "myapp"
    type: "systemd"
    service: "myapp.service"
    env_file: "/opt/myapp/.env"
    health_check:
      type: "http"
      url: "http://localhost:3000/health"
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.VM.Name != "test-vm" {
		t.Errorf("vm.name = %q, want %q", cfg.VM.Name, "test-vm")
	}
	if cfg.VM.Port != 9000 {
		t.Errorf("vm.port = %d, want 9000", cfg.VM.Port)
	}
	if len(cfg.Apps) != 1 {
		t.Fatalf("len(apps) = %d, want 1", len(cfg.Apps))
	}
	if cfg.Apps[0].Name != "myapp" {
		t.Errorf("apps[0].name = %q, want %q", cfg.Apps[0].Name, "myapp")
	}
	if cfg.Apps[0].HealthCheck.Type != "http" {
		t.Errorf("health_check.type = %q, want %q", cfg.Apps[0].HealthCheck.Type, "http")
	}
}

func TestLoad_DefaultPort(t *testing.T) {
	path := writeTemp(t, `
vm:
  name: "test-vm"
  control_plane_url: "http://localhost:8080"
  auth_token: "secret"
apps: []
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.VM.Port != 9000 {
		t.Errorf("default port = %d, want 9000", cfg.VM.Port)
	}
}

func TestLoad_MissingName(t *testing.T) {
	path := writeTemp(t, `
vm:
  control_plane_url: "http://localhost:8080"
  auth_token: "secret"
apps: []
`)
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for missing vm.name, got nil")
	}
}

func TestLoad_MissingAuthToken(t *testing.T) {
	path := writeTemp(t, `
vm:
  name: "test-vm"
  control_plane_url: "http://localhost:8080"
apps: []
`)
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for missing auth_token, got nil")
	}
}

func TestLoad_InvalidAppType(t *testing.T) {
	path := writeTemp(t, `
vm:
  name: "test-vm"
  control_plane_url: "http://localhost:8080"
  auth_token: "secret"
apps:
  - name: "myapp"
    type: "pm2"
    service: "myapp"
`)
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for invalid app type, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	if _, err := Load("/nonexistent/path/agent.yaml"); err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
