package health

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vidya381/vm-monitor/agent/internal/config"
)

func TestCheckHTTP_Healthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	result := Check(config.HealthCheck{Type: "http", URL: srv.URL})
	if !result.Healthy {
		t.Errorf("expected healthy, got: %s", result.Message)
	}
}

func TestCheckHTTP_Unhealthy_5xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	result := Check(config.HealthCheck{Type: "http", URL: srv.URL})
	if result.Healthy {
		t.Error("expected unhealthy for 500 response")
	}
}

func TestCheckHTTP_Unhealthy_ConnectionRefused(t *testing.T) {
	result := Check(config.HealthCheck{Type: "http", URL: "http://127.0.0.1:1"})
	if result.Healthy {
		t.Error("expected unhealthy for refused connection")
	}
	if result.Message == "" {
		t.Error("expected error message")
	}
}

func TestCheckCommand_Healthy(t *testing.T) {
	result := Check(config.HealthCheck{Type: "command", Cmd: "true"})
	if !result.Healthy {
		t.Errorf("expected healthy, got: %s", result.Message)
	}
}

func TestCheckCommand_Unhealthy(t *testing.T) {
	result := Check(config.HealthCheck{Type: "command", Cmd: "false"})
	if result.Healthy {
		t.Error("expected unhealthy for failing command")
	}
}

func TestCheck_NoHealthCheck(t *testing.T) {
	result := Check(config.HealthCheck{})
	if !result.Healthy {
		t.Error("no health check configured should return healthy")
	}
}
