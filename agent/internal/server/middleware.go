package server

import (
	"net/http"
	"strings"

	"github.com/vidya381/vm-monitor/agent/internal/env"
)

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// /health is exempt from auth
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !env.TokensMatch(token, s.cfg.VM.AuthToken) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
