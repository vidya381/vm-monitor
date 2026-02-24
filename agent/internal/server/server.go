package server

import (
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vidya381/vm-monitor/agent/internal/config"
)

type Server struct {
	cfg     *config.Config
	cfgPath string
	mu      sync.RWMutex
	router  *chi.Mux
}

func New(cfg *config.Config, cfgPath string) *Server {
	s := &Server{cfg: cfg, cfgPath: cfgPath}
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
	r.Post("/apps", s.handleAddApp)
	r.Get("/apps/{id}/status", s.handleAppStatus)
	r.Get("/apps/{id}/logs", s.handleAppLogs)
	r.Get("/apps/{id}/logs/stream", s.handleLogStream)
	r.Get("/apps/{id}/env/files", s.handleListEnvFiles)
	r.Get("/apps/{id}/env", s.handleGetEnv)
	r.Put("/apps/{id}/env", s.handlePutEnv)
	r.Post("/apps/{id}/restart", s.handleRestart)
	r.Post("/apps/{id}/deploy", s.handleDeploy)
	r.Get("/apps/{id}/metrics", s.handleMetrics)
	r.Get("/system/metrics", s.handleSystemMetrics)

	return r
}
