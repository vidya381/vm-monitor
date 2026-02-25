package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/vidya381/vm-monitor/agent/internal/config"
	"github.com/vidya381/vm-monitor/agent/internal/server"
)

var version = "dev"

func main() {
	cfgPath := flag.String("config", "/etc/vm-monitor/agent.yaml", "path to agent.yaml")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	slog.Info("starting vm-monitor agent",
		"version", version,
		"vm", cfg.VM.Name,
		"port", cfg.VM.Port,
		"apps", len(cfg.Apps),
	)

	srv := server.New(cfg, *cfgPath)
	addr := fmt.Sprintf("0.0.0.0:%d", cfg.VM.Port)

	if err := register(cfg); err != nil {
		slog.Warn("failed to register with control plane, will retry in background", "error", err)
		go retryRegister(cfg)
	}

	slog.Info("listening", "addr", addr)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

// registerRequest mirrors the control plane's model.RegisterRequest.
type registerRequest struct {
	Name      string       `json:"name"`
	Address   string       `json:"address"`
	AuthToken string       `json:"auth_token"`
	Labels    []string     `json:"labels"`
	Apps      []appInput   `json:"apps"`
}

type appInput struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Environment string    `json:"environment"`
	Config      appConfig `json:"config"`
}

type appConfig struct {
	Service   string `json:"service,omitempty"`
	Container string `json:"container,omitempty"`
	EnvFile   string `json:"env_file,omitempty"`
	DeployDir string `json:"deploy_dir,omitempty"`
}

func retryRegister(cfg *config.Config) {
	for attempt := 1; ; attempt++ {
		time.Sleep(30 * time.Second)
		if err := register(cfg); err != nil {
			slog.Warn("registration retry failed", "attempt", attempt, "error", err)
			continue
		}
		return
	}
}

func register(cfg *config.Config) error {
	apps := make([]appInput, len(cfg.Apps))
	for i, a := range cfg.Apps {
		apps[i] = appInput{
			Name:        a.Name,
			Type:        a.Type,
			Environment: a.Environment,
			Config: appConfig{
				Service:   a.Service,
				Container: a.Container,
				EnvFile:   a.EnvFile,
				DeployDir: a.DeployDir,
			},
		}
	}

	payload := registerRequest{
		Name:      cfg.VM.Name,
		Address:   cfg.VM.Address,
		AuthToken: cfg.VM.AuthToken,
		Labels:    cfg.VM.Labels,
		Apps:      apps,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling payload: %w", err)
	}

	url := cfg.VM.ControlPlaneURL + "/vms/register"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.VM.ControlPlaneAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("calling control plane: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("control plane returned %d", resp.StatusCode)
	}

	slog.Info("registered with control plane", "vm", cfg.VM.Name, "apps", len(apps))
	return nil
}
