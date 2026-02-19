package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/vidya381/vm-monitor/agent/internal/config"
	"github.com/vidya381/vm-monitor/agent/internal/server"
)

func main() {
	cfgPath := "/etc/vm-monitor/agent.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	slog.Info("starting vm-monitor agent",
		"vm", cfg.VM.Name,
		"port", cfg.VM.Port,
		"apps", len(cfg.Apps),
	)

	srv := server.New(cfg)
	addr := fmt.Sprintf("0.0.0.0:%d", cfg.VM.Port)

	slog.Info("listening", "addr", addr)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
