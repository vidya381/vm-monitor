package handler

import (
	"encoding/json"
	"net/http"

	"github.com/vidya381/vm-monitor/api/internal/agentclient"
	"github.com/vidya381/vm-monitor/api/internal/db"
)

type Handler struct {
	vms    *db.VMStore
	apps   *db.AppStore
	audit  *db.AuditStore
	agent  *agentclient.Client
}

func New(vms *db.VMStore, apps *db.AppStore, audit *db.AuditStore, agent *agentclient.Client) *Handler {
	return &Handler{vms: vms, apps: apps, audit: audit, agent: agent}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
