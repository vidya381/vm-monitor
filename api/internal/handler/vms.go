package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/vidya381/vm-monitor/api/internal/db"
	"github.com/vidya381/vm-monitor/api/internal/model"
)

func (h *Handler) RegisterVM(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Address == "" || req.AuthToken == "" {
		writeError(w, http.StatusBadRequest, "name, address, and auth_token are required")
		return
	}
	if len(req.Name) > 255 {
		writeError(w, http.StatusBadRequest, "name must be 255 characters or fewer")
		return
	}
	if len(req.AuthToken) < 16 {
		writeError(w, http.StatusBadRequest, "auth_token must be at least 16 characters")
		return
	}
	if u, err := url.Parse(req.Address); err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		writeError(w, http.StatusBadRequest, "address must be a valid http/https URL")
		return
	}

	vm, err := h.vms.Upsert(r.Context(), model.VM{
		Name:      req.Name,
		Address:   req.Address,
		AuthToken: req.AuthToken,
		Labels:    req.Labels,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to register VM")
		return
	}

	for _, app := range req.Apps {
		if err := h.apps.Upsert(r.Context(), vm.ID, app); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to register app: "+app.Name)
			return
		}
	}

	writeJSON(w, http.StatusOK, vm)
}

func (h *Handler) ListVMs(w http.ResponseWriter, r *http.Request) {
	vms, err := h.vms.GetAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list VMs")
		return
	}
	if vms == nil {
		vms = []model.VM{}
	}
	writeJSON(w, http.StatusOK, vms)
}

func (h *Handler) GetVM(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid VM id")
		return
	}

	vm, err := h.vms.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			writeError(w, http.StatusNotFound, "VM not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get VM")
		return
	}
	writeJSON(w, http.StatusOK, vm)
}
