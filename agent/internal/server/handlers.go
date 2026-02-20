package server

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/vidya381/vm-monitor/agent/internal/config"
	agentenv "github.com/vidya381/vm-monitor/agent/internal/env"
	"github.com/vidya381/vm-monitor/agent/internal/health"
	"github.com/vidya381/vm-monitor/agent/internal/systemd"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleListApps(w http.ResponseWriter, r *http.Request) {
	type appResponse struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Type   string `json:"type"`
		Status string `json:"status"`
	}

	apps := make([]appResponse, 0, len(s.cfg.Apps))
	for _, app := range s.cfg.Apps {
		apps = append(apps, appResponse{
			ID:     app.Name,
			Name:   app.Name,
			Type:   app.Type,
			Status: string(appStatus(&app)),
		})
	}
	writeJSON(w, http.StatusOK, apps)
}

func (s *Server) handleAppStatus(w http.ResponseWriter, r *http.Request) {
	app, ok := s.findApp(w, r)
	if !ok {
		return
	}

	type statusResponse struct {
		Status      string        `json:"status"`
		HealthCheck health.Result `json:"health_check"`
	}

	svcStatus := systemd.AppStatus(app.Service)
	hc := health.Check(app.HealthCheck)

	writeJSON(w, http.StatusOK, statusResponse{
		Status:      string(appStatus(app)),
		HealthCheck: hc,
	})
	_ = svcStatus
}

func (s *Server) handleAppLogs(w http.ResponseWriter, r *http.Request) {
	app, ok := s.findApp(w, r)
	if !ok {
		return
	}

	tail := 200
	if t := r.URL.Query().Get("tail"); t != "" {
		if n, err := strconv.Atoi(t); err == nil && n > 0 {
			tail = n
		}
	}
	cursor := r.URL.Query().Get("cursor")

	result, err := systemd.Logs(app.Service, tail, cursor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleGetEnv(w http.ResponseWriter, r *http.Request) {
	app, ok := s.findApp(w, r)
	if !ok {
		return
	}
	if app.EnvFile == "" {
		writeJSON(w, http.StatusOK, map[string]agentenv.EnvVar{})
		return
	}

	vars, err := agentenv.Parse(app.EnvFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, vars)
}

func (s *Server) handlePutEnv(w http.ResponseWriter, r *http.Request) {
	app, ok := s.findApp(w, r)
	if !ok {
		return
	}

	var updates map[string]string
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Read existing env and merge updates on top so unedited keys (especially
	// masked secrets) are preserved in the file even if not included in the PUT.
	current, err := agentenv.ParseRaw(app.EnvFile)
	if err != nil && !os.IsNotExist(err) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if current == nil {
		current = make(map[string]string)
	}
	for k, v := range updates {
		current[k] = v
	}

	if err := agentenv.Write(app.EnvFile, current); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.URL.Query().Get("restart") == "true" {
		if err := systemd.Restart(app.Service); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleRestart(w http.ResponseWriter, r *http.Request) {
	app, ok := s.findApp(w, r)
	if !ok {
		return
	}
	if err := systemd.Restart(app.Service); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) findApp(w http.ResponseWriter, r *http.Request) (*config.AppConfig, bool) {
	id := chi.URLParam(r, "id")
	for i := range s.cfg.Apps {
		if s.cfg.Apps[i].Name == id {
			return &s.cfg.Apps[i], true
		}
	}
	http.Error(w, "app not found", http.StatusNotFound)
	return nil, false
}

// appStatus returns the combined systemd + health check status for an app.
// Running  = service active AND health check passes (or no health check)
// Unhealthy = service active but health check fails
// Stopped   = service not active
func appStatus(app *config.AppConfig) systemd.Status {
	svcStatus := systemd.AppStatus(app.Service)
	if svcStatus != systemd.StatusRunning {
		return svcStatus
	}
	if app.HealthCheck.Type != "" {
		if !health.Check(app.HealthCheck).Healthy {
			return systemd.StatusUnhealthy
		}
	}
	return systemd.StatusRunning
}
