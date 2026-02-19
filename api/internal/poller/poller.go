package poller

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/vidya381/vm-monitor/api/internal/agentclient"
	"github.com/vidya381/vm-monitor/api/internal/db"
	"github.com/vidya381/vm-monitor/api/internal/model"
)

// Start launches a background goroutine that polls every registered agent on
// the given interval, updating VM status and app last_status in the database.
func Start(ctx context.Context, vms *db.VMStore, apps *db.AppStore, client *agentclient.Client, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				poll(ctx, vms, apps, client)
			}
		}
	}()
}

func poll(ctx context.Context, vms *db.VMStore, apps *db.AppStore, client *agentclient.Client) {
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
			pollVM(ctx, v, vms, apps, client)
		}(vm)
	}
	wg.Wait()
}

func pollVM(ctx context.Context, vm model.VM, vms *db.VMStore, apps *db.AppStore, client *agentclient.Client) {
	// Short per-agent timeout so a slow agent doesn't block the rest.
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
	}

	slog.Debug("poller: polled", "vm", vm.Name, "apps", len(agentApps))
}

