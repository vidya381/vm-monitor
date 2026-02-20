package middleware

import (
	"net/http"
	"sync"
	"time"
)

// ipWindow tracks a fixed-window request count for a single IP.
type ipWindow struct {
	mu      sync.Mutex
	count   int
	resetAt time.Time
}

// RateLimit returns middleware that limits each IP to limit requests per window.
// Excess requests receive 429 Too Many Requests.
func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	var limiters sync.Map // map[string]*ipWindow

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			val, _ := limiters.LoadOrStore(ip, &ipWindow{resetAt: time.Now().Add(window)})
			win := val.(*ipWindow)

			win.mu.Lock()
			now := time.Now()
			if now.After(win.resetAt) {
				win.count = 0
				win.resetAt = now.Add(window)
			}
			win.count++
			over := win.count > limit
			win.mu.Unlock()

			if over {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
