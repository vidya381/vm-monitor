package health

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/vidya381/vm-monitor/agent/internal/config"
)

type Result struct {
	Healthy bool   `json:"healthy"`
	Message string `json:"message,omitempty"`
}

// Check runs the configured health check for an app.
// Returns a healthy result if no health check is configured.
func Check(cfg config.HealthCheck) Result {
	switch cfg.Type {
	case "http":
		return checkHTTP(cfg.URL)
	case "command":
		return checkCommand(cfg.Cmd)
	default:
		return Result{Healthy: true}
	}
}

func checkHTTP(url string) Result {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{Healthy: false, Message: fmt.Sprintf("bad url: %v", err)}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Result{Healthy: false, Message: fmt.Sprintf("request failed: %v", err)}
	}
	resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return Result{Healthy: true}
	}
	return Result{Healthy: false, Message: fmt.Sprintf("status %d", resp.StatusCode)}
}

func checkCommand(cmd string) Result {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := exec.CommandContext(ctx, "sh", "-c", cmd).Run(); err != nil {
		return Result{Healthy: false, Message: err.Error()}
	}
	return Result{Healthy: true}
}
