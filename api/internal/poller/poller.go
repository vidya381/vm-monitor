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

// incident tracks an ongoing outage for one app.
type incident struct {
	firing  bool
	since   time.Time
}

// Start launches a background goroutine that polls every registered agent on
// the given interval, updating VM status and app last_status in the database.
func Start(ctx context.Context, vms *db.VMStore, apps *db.AppStore, client *agentclient.Client, notifier *notify.Notifier, interval time.Duration) {
	incidents := make(map[uuid.UUID]*incident)
	var mu sync.Mutex

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				poll(ctx, vms, apps, client, notifier, incidents, &mu)
			}
		}
	}()
}

func poll(ctx context.Context, vms *db.VMStore, apps *db.AppStore, client *agentclient.Client, notifier *notify.Notifier, incidents map[uuid.UUID]*incident, mu *sync.Mutex) {
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
			pollVM(ctx, v, vms, apps, client, notifier, incidents, mu)
		}(vm)
	}
	wg.Wait()
}

func pollVM(ctx context.Context, vm model.VM, vms *db.VMStore, apps *db.AppStore, client *agentclient.Client, notifier *notify.Notifier, incidents map[uuid.UUID]*incident, mu *sync.Mutex) {
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

		// Look up the app ID for incident tracking.
		app, err := apps.GetByVMAndName(pollCtx, vm.ID, a.Name)
		if err != nil {
			continue
		}

		handleIncident(app.ID, vm.Name, a.Name, a.Status, notifier, incidents, mu)
	}

	slog.Debug("poller: polled", "vm", vm.Name, "apps", len(agentApps))
}

// handleIncident fires notifications on status transitions.
// Down alert fires once when an app first becomes unhealthy/stopped.
// Recovery fires once when it returns to running.
func handleIncident(appID uuid.UUID, vmName, appName, status string, notifier *notify.Notifier, incidents map[uuid.UUID]*incident, mu *sync.Mutex) {
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
			Type:    notify.EventDown,
			VMName:  vmName,
			AppName: appName,
		}
		if err := notifier.Send(event); err != nil {
			slog.Warn("poller: failed to send down notification", "app", appName, "error", err)
		} else {
			slog.Info("poller: sent down notification", "vm", vmName, "app", appName)
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
