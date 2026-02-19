package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"github.com/vidya381/vm-monitor/api/internal/agentclient"
	"github.com/vidya381/vm-monitor/api/internal/db"
	"github.com/vidya381/vm-monitor/api/internal/handler"
	apimw "github.com/vidya381/vm-monitor/api/internal/middleware"
	"github.com/vidya381/vm-monitor/api/internal/poller"
)

func main() {
	dsn := mustEnv("DATABASE_URL")
	apiKey := mustEnv("API_KEY")
	allowedOrigins := mustEnv("ALLOWED_ORIGINS")

	ctx := context.Background()

	pool, err := db.Connect(ctx, dsn)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, pool); err != nil {
		slog.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	vms := db.NewVMStore(pool)
	apps := db.NewAppStore(pool)
	audit := db.NewAuditStore(pool)
	agent := agentclient.New()

	h := handler.New(vms, apps, audit, agent)

	poller.Start(ctx, vms, apps, agent, 30*time.Second)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{allowedOrigins},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	})

	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(c.Handler)

	r.Get("/health", h.Health)

	r.Group(func(r chi.Router) {
		r.Use(apimw.APIKey(apiKey))

		r.Post("/vms/register", h.RegisterVM)
		r.Get("/vms", h.ListVMs)
		r.Get("/vms/{id}", h.GetVM)

		r.Get("/apps", h.ListApps)
		r.Get("/apps/{id}", h.GetApp)
		r.Get("/apps/{id}/logs", h.GetAppLogs)
		r.Get("/apps/{id}/env", h.GetAppEnv)
		r.Put("/apps/{id}/env", h.PutAppEnv)
		r.Post("/apps/{id}/restart", h.RestartApp)
		r.Get("/apps/{id}/audit", h.GetAppAudit)

		r.Get("/audit", h.ListAudit)
	})

	addr := ":8080"
	slog.Info("starting vm-monitor api", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		slog.Error("required env var not set", "key", key)
		os.Exit(1)
	}
	return v
}
