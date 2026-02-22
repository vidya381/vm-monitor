package poller

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vidya381/vm-monitor/api/internal/agentclient"
	"github.com/vidya381/vm-monitor/api/internal/db"
	"github.com/vidya381/vm-monitor/api/internal/model"
	"github.com/vidya381/vm-monitor/api/internal/notify"
)

const (
	restartWindowDuration = 10 * time.Minute
	maxRestartsPerWindow  = 3
)

// incident tracks an ongoing outage for one app.
type incident struct {
	firing  bool
	since   time.Time
}

// restartTracker tracks consecutive auto-restart attempts for flap protection.
type restartTracker struct {
	count       int
	windowStart time.Time
}

// Start launches a background goroutine that polls every registered agent on
// the given interval, updating VM status and app last_status in the database.
func Start(ctx context.Context, vms *db.VMStore, apps *db.AppStore, audit *db.AuditStore, client *agentclient.Client, notifier *notify.Notifier, interval time.Duration) {
	incidents := make(map[uuid.UUID]*incident)
	restarts := make(map[uuid.UUID]*restartTracker)
	var mu sync.Mutex

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				poll(ctx, vms, apps, audit, client, notifier, incidents, restarts, &mu)
			}
		}
	}()
}

func poll(ctx context.Context, vms *db.VMStore, apps *db.AppStore, audit *db.AuditStore, client *agentclient.Client, notifier *notify.Notifier, incidents map[uuid.UUID]*incident, restarts map[uuid.UUID]*restartTracker, mu *sync.Mutex) {
	allVMs, err := vms.GetAll(ctx)
	if err != nil {
		slog.Error("poller: failed to list VMs", "error", err)
		return
	}

	var wg sync.WaitGroup
	for _, vm := range allVMs {
		wg.Add(1)
		go func(v model.VM) {
			defer wg.Done()
			pollVM(ctx, v, vms, apps, audit, client, notifier, incidents, restarts, mu)
		}(vm)
	}
	wg.Wait()
}

func pollVM(ctx context.Context, vm model.VM, vms *db.VMStore, apps *db.AppStore, audit *db.AuditStore, client *agentclient.Client, notifier *notify.Notifier, incidents map[uuid.UUID]*incident, restarts map[uuid.UUID]*restartTracker, mu *sync.Mutex) {
	pollCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	now := time.Now()
	agentApps, err := client.GetApps(vm.Address, vm.AuthToken)
	if err != nil {
		slog.Warn("poller: agent unreachable", "vm", vm.Name, "error", err)
		vms.UpdateStatus(pollCtx, vm.ID, "unreachable", now)
		return
	}

	vms.UpdateStatus(pollCtx, vm.ID, "online", now)

	for _, a := range agentApps {
		if err := apps.UpdateStatus(pollCtx, vm.ID, a.Name, a.Status, now); err != nil {
			slog.Warn("poller: failed to update app status", "vm", vm.Name, "app", a.Name, "error", err)
		}

		app, err := apps.GetByVMAndName(pollCtx, vm.ID, a.Name)
		if err != nil {
			continue
		}

		isDown := a.Status != "running"
		autoRestarted := false
		if isDown && app.Config.AutoRestart {
			autoRestarted = tryAutoRestart(pollCtx, app, a.Status, client, apps, audit, restarts, mu)
		}

		handleIncident(app.ID, vm.Name, a.Name, a.Status, autoRestarted, notifier, incidents, mu)
	}

	slog.Debug("poller: polled", "vm", vm.Name, "apps", len(agentApps))
}

// tryAutoRestart attempts to restart a stopped/unhealthy app, subject to flap
// protection (max 3 restarts per 10 minutes). Returns true if restart was called.
func tryAutoRestart(ctx context.Context, app *model.App, status string, client *agentclient.Client, apps *db.AppStore, audit *db.AuditStore, restarts map[uuid.UUID]*restartTracker, mu *sync.Mutex) bool {
	mu.Lock()
	rt, ok := restarts[app.ID]
	if !ok {
		rt = &restartTracker{}
		restarts[app.ID] = rt
	}
	// Reset window if expired.
	if !rt.windowStart.IsZero() && time.Since(rt.windowStart) > restartWindowDuration {
		rt.count = 0
		rt.windowStart = time.Time{}
	}
	if rt.count >= maxRestartsPerWindow {
		mu.Unlock()
		slog.Warn("poller: auto-restart flap protection, skipping", "vm", app.VMName, "app", app.Name, "count", rt.count)
		return false
	}
	if rt.windowStart.IsZero() {
		rt.windowStart = time.Now()
	}
	rt.count++
	mu.Unlock()

	if err := client.Restart(ctx, app.VMAddress, app.VMAuthToken, app.Name); err != nil {
		slog.Warn("poller: auto-restart failed", "vm", app.VMName, "app", app.Name, "error", err)
		return false
	}

	apps.UpdateLastRestarted(ctx, app.ID, time.Now())
	audit.Create(ctx, app.ID, "auto_restart", map[string]any{"reason": status})
	slog.Info("poller: auto-restarted app", "vm", app.VMName, "app", app.Name)
	return true
}

// handleIncident fires notifications on status transitions.
// Down alert fires once when an app first becomes unhealthy/stopped.
// Recovery fires once when it returns to running.
func handleIncident(appID uuid.UUID, vmName, appName, status string, autoRestarted bool, notifier *notify.Notifier, incidents map[uuid.UUID]*incident, mu *sync.Mutex) {
	if !notifier.Enabled() {
		return
	}

	mu.Lock()
	inc, ok := incidents[appID]
	if !ok {
		inc = &incident{}
		incidents[appID] = inc
	}
	mu.Unlock()

	isDown := status != "running"

	if isDown && !inc.firing {
		// Transition: healthy → down
		inc.firing = true
		inc.since = time.Now()
		event := notify.Event{
			Type:          notify.EventDown,
			VMName:        vmName,
			AppName:       appName,
			AutoRestarted: autoRestarted,
		}
		if err := notifier.Send(event); err != nil {
			slog.Warn("poller: failed to send down notification", "app", appName, "error", err)
		} else {
			slog.Info("poller: sent down notification", "vm", vmName, "app", appName, "auto_restarted", autoRestarted)
		}
	} else if !isDown && inc.firing {
		// Transition: down → healthy
		duration := time.Since(inc.since)
		inc.firing = false
		inc.since = time.Time{}
		event := notify.Event{
			Type:     notify.EventRecovered,
			VMName:   vmName,
			AppName:  appName,
			Duration: duration,
		}
		if err := notifier.Send(event); err != nil {
			slog.Warn("poller: failed to send recovery notification", "app", appName, "error", err)
		} else {
			slog.Info("poller: sent recovery notification", "vm", vmName, "app", appName, "downtime", duration)
		}
	}
}
