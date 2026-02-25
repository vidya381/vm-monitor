package poller

import (
	"context"
	"log/slog"
	"time"

	"github.com/vidya381/vm-monitor/api/internal/db"
)

// StartPurgeJob runs a daily background goroutine that deletes status_history
// rows older than 90 days.
func StartPurgeJob(ctx context.Context, history *db.StatusHistoryStore) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(24 * time.Hour):
				if err := history.Purge(ctx); err != nil {
					slog.Warn("purge: failed to clean up status_history", "error", err)
				} else {
					slog.Info("purge: status_history cleanup completed")
				}
			}
		}
	}()
}
