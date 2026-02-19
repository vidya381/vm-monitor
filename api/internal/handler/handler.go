package handler

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// VM handlers — stubs to be implemented in M3
func (h *Handler) RegisterVM(w http.ResponseWriter, r *http.Request)  {}
func (h *Handler) ListVMs(w http.ResponseWriter, r *http.Request)     {}
func (h *Handler) GetVM(w http.ResponseWriter, r *http.Request)       {}

// App handlers — stubs to be implemented in M3
func (h *Handler) ListApps(w http.ResponseWriter, r *http.Request)    {}
func (h *Handler) GetApp(w http.ResponseWriter, r *http.Request)      {}
func (h *Handler) GetAppLogs(w http.ResponseWriter, r *http.Request)  {}
func (h *Handler) GetAppEnv(w http.ResponseWriter, r *http.Request)   {}
func (h *Handler) PutAppEnv(w http.ResponseWriter, r *http.Request)   {}
func (h *Handler) RestartApp(w http.ResponseWriter, r *http.Request)  {}
func (h *Handler) GetAppAudit(w http.ResponseWriter, r *http.Request) {}
func (h *Handler) ListAudit(w http.ResponseWriter, r *http.Request)   {}
