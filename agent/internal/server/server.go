package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vidya381/vm-monitor/agent/internal/config"
)

type Server struct {
	cfg    *config.Config
	router *chi.Mux
}

func New(cfg *config.Config) *Server {
	s := &Server{cfg: cfg}
	s.router = s.buildRouter()
	return s
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) buildRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(s.authMiddleware)

	r.Get("/health", s.handleHealth)
	r.Get("/apps", s.handleListApps)
	r.Get("/apps/{id}/status", s.handleAppStatus)
	r.Get("/apps/{id}/logs", s.handleAppLogs)
	r.Get("/apps/{id}/env/files", s.handleListEnvFiles)
	r.Get("/apps/{id}/env", s.handleGetEnv)
	r.Put("/apps/{id}/env", s.handlePutEnv)
	r.Post("/apps/{id}/restart", s.handleRestart)

	return r
}
