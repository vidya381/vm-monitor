package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/vidya381/vm-monitor/agent/internal/config"
	"github.com/vidya381/vm-monitor/agent/internal/docker"
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

	var result *systemd.LogResult
	var err error
	if app.Type == "docker" {
		result, err = docker.Logs(app.Container, tail, cursor)
	} else {
		result, err = systemd.Logs(app.Service, tail, cursor)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleListEnvFiles(w http.ResponseWriter, r *http.Request) {
	app, ok := s.findApp(w, r)
	if !ok {
		return
	}
	all := app.AllEnvFiles()
	names := make([]string, len(all))
	for i, f := range all {
		names[i] = filepath.Base(f)
	}
	writeJSON(w, http.StatusOK, names)
}

func (s *Server) handleGetEnv(w http.ResponseWriter, r *http.Request) {
	app, ok := s.findApp(w, r)
	if !ok {
		return
	}
	path, err := resolveEnvFile(app, r.URL.Query().Get("file"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if path == "" {
		writeJSON(w, http.StatusOK, map[string]agentenv.EnvVar{})
		return
	}
	vars, err := agentenv.Parse(path)
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
	path, err := resolveEnvFile(app, r.URL.Query().Get("file"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if path == "" {
		http.Error(w, "no env file configured", http.StatusBadRequest)
		return
	}

	var updates map[string]string
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Read existing env and merge updates on top so unedited keys (especially
	// masked secrets) are preserved in the file even if not included in the PUT.
	current, err := agentenv.ParseRaw(path)
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

	if err := agentenv.Write(path, current); err != nil {
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

// resolveEnvFile returns the full path for the requested env file.
// If fileParam is empty, defaults to the first configured file.
func resolveEnvFile(app *config.AppConfig, fileParam string) (string, error) {
	files := app.AllEnvFiles()
	if len(files) == 0 {
		return "", nil
	}
	if fileParam == "" {
		return files[0], nil
	}
	for _, f := range files {
		if filepath.Base(f) == fileParam {
			return f, nil
		}
	}
	return "", fmt.Errorf("env file %q not configured for this app", fileParam)
}

func (s *Server) handleRestart(w http.ResponseWriter, r *http.Request) {
	app, ok := s.findApp(w, r)
	if !ok {
		return
	}
	var err error
	if app.Type == "docker" {
		err = docker.Restart(app.Container)
	} else {
		err = systemd.Restart(app.Service)
	}
	if err != nil {
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

// appStatus returns the combined status for an app (systemd or docker) plus
// the result of an optional health check.
// Running   = process active AND health check passes (or no health check)
// Unhealthy = process active but health check fails
// Stopped   = process not active
func appStatus(app *config.AppConfig) systemd.Status {
	var base systemd.Status
	if app.Type == "docker" {
		base = docker.AppStatus(app.Container)
	} else {
		base = systemd.AppStatus(app.Service)
	}
	if base != systemd.StatusRunning {
		return base
	}
	if app.HealthCheck.Type != "" {
		if !health.Check(app.HealthCheck).Healthy {
			return systemd.StatusUnhealthy
		}
	}
	return systemd.StatusRunning
}
