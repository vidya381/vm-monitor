package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/vidya381/vm-monitor/api/internal/db"
	"github.com/vidya381/vm-monitor/api/internal/model"
)

func (h *Handler) ListApps(w http.ResponseWriter, r *http.Request) {
	var apps []model.App
	var err error

	if vmIDStr := r.URL.Query().Get("vm_id"); vmIDStr != "" {
		vmID, parseErr := uuid.Parse(vmIDStr)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, "invalid vm_id")
			return
		}
		apps, err = h.apps.GetByVMID(r.Context(), vmID)
	} else {
		apps, err = h.apps.GetAll(r.Context())
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list apps")
		return
	}
	if apps == nil {
		apps = []model.App{}
	}
	writeJSON(w, http.StatusOK, apps)
}

func (h *Handler) GetApp(w http.ResponseWriter, r *http.Request) {
	app, ok := h.resolveApp(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, app)
}

func (h *Handler) GetAppLogs(w http.ResponseWriter, r *http.Request) {
	app, ok := h.resolveApp(w, r)
	if !ok {
		return
	}
	agentURL := fmt.Sprintf("%s/apps/%s/logs", app.VMAddress, app.Name)
	h.agent.ProxyRequest(w, r, agentURL, app.VMAuthToken)
}

func (h *Handler) StreamAppLogs(w http.ResponseWriter, r *http.Request) {
	app, ok := h.resolveApp(w, r)
	if !ok {
		return
	}

	agentURL := fmt.Sprintf("%s/apps/%s/logs/stream", app.VMAddress, app.Name)
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, agentURL, nil)
	if err != nil {
		http.Error(w, "failed to build request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+app.VMAuthToken)

	// Use a client with no timeout — SSE connections are long-lived.
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		http.Error(w, "agent unreachable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}

	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
			flusher.Flush()
		}
		if err != nil {
			if err != io.EOF {
				// client disconnected or agent closed — normal exit
			}
			return
		}
	}
}

func (h *Handler) GetAppEnvFiles(w http.ResponseWriter, r *http.Request) {
	app, ok := h.resolveApp(w, r)
	if !ok {
		return
	}
	agentURL := fmt.Sprintf("%s/apps/%s/env/files", app.VMAddress, app.Name)
	h.agent.ProxyRequest(w, r, agentURL, app.VMAuthToken)
}

func (h *Handler) GetAppEnv(w http.ResponseWriter, r *http.Request) {
	app, ok := h.resolveApp(w, r)
	if !ok {
		return
	}
	agentURL := fmt.Sprintf("%s/apps/%s/env", app.VMAddress, app.Name)
	h.agent.ProxyRequest(w, r, agentURL, app.VMAuthToken)
}

func (h *Handler) PutAppEnv(w http.ResponseWriter, r *http.Request) {
	app, ok := h.resolveApp(w, r)
	if !ok {
		return
	}

	// Capture request body to create audit log diff.
	var newVars map[string]string
	if err := json.NewDecoder(r.Body).Decode(&newVars); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	agentURL := fmt.Sprintf("%s/apps/%s/env", app.VMAddress, app.Name)

	// Rebuild request with decoded body so we can proxy it cleanly.
	encoded, _ := json.Marshal(newVars)
	r2, _ := http.NewRequestWithContext(r.Context(), http.MethodPut, agentURL, bytes.NewReader(encoded))
	r2.URL.RawQuery = r.URL.RawQuery
	status := h.agent.ProxyRequest(w, r2, agentURL, app.VMAuthToken)

	if status >= 200 && status < 300 {
		h.audit.Create(r.Context(), app.ID, "env_update", map[string]any{
			"keys_updated": len(newVars),
			"restart":      r.URL.Query().Get("restart") == "true",
		})
	}
}

func (h *Handler) RestartApp(w http.ResponseWriter, r *http.Request) {
	app, ok := h.resolveApp(w, r)
	if !ok {
		return
	}

	agentURL := fmt.Sprintf("%s/apps/%s/restart", app.VMAddress, app.Name)
	status := h.agent.ProxyRequest(w, r, agentURL, app.VMAuthToken)

	if status >= 200 && status < 300 {
		h.apps.UpdateLastRestarted(r.Context(), app.ID, time.Now())
		h.audit.Create(r.Context(), app.ID, "restart", nil)
	}
}

func (h *Handler) GetAppAudit(w http.ResponseWriter, r *http.Request) {
	app, ok := h.resolveApp(w, r)
	if !ok {
		return
	}
	logs, err := h.audit.GetByAppID(r.Context(), app.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch audit logs")
		return
	}
	if logs == nil {
		logs = []model.AuditLog{}
	}
	writeJSON(w, http.StatusOK, logs)
}

func (h *Handler) ListAudit(w http.ResponseWriter, r *http.Request) {
	logs, err := h.audit.GetAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch audit logs")
		return
	}
	if logs == nil {
		logs = []model.AuditLog{}
	}
	writeJSON(w, http.StatusOK, logs)
}

func (h *Handler) resolveApp(w http.ResponseWriter, r *http.Request) (*model.App, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid app id")
		return nil, false
	}
	app, err := h.apps.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(w, http.StatusNotFound, "app not found")
			return nil, false
		}
		writeError(w, http.StatusInternalServerError, "failed to get app")
		return nil, false
	}
	return app, true
}

