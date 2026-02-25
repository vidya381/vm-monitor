"use client";

import { useState, useEffect } from "react";
import { AppCard } from "@/components/app-card";
import { App } from "@/lib/types";

export default function OverviewPage() {
  const [apps, setApps] = useState<App[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [lastRefresh, setLastRefresh] = useState(new Date());
  const [uptimes, setUptimes] = useState<Record<string, number | null>>({});

  const fetchApps = async () => {
    try {
      const res = await fetch("/api/apps");
      if (!res.ok) throw new Error(`status ${res.status}`);
      const data: App[] = await res.json();
      setApps(data);
      setLastRefresh(new Date());
      setError("");

      // Fetch uptime for each app in parallel (best-effort).
      const results = await Promise.all(
        data.map(async (app) => {
          try {
            const r = await fetch(`/api/apps/${app.id}/uptime?days=30`);
            if (!r.ok) return [app.id, null] as const;
            const u = await r.json();
            return [app.id, u?.uptime_pct ?? null] as const;
          } catch {
            return [app.id, null] as const;
          }
        })
      );
      setUptimes(Object.fromEntries(results));
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchApps();
    const interval = setInterval(fetchApps, 30000);
    return () => clearInterval(interval);
  }, []);

  if (loading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        {[...Array(3)].map((_, i) => (
          <div key={i} className="h-32 rounded-lg border border-border bg-surface animate-pulse" />
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-text-primary">Apps</h1>
          <p className="text-sm text-text-secondary">
            {apps.length} app{apps.length !== 1 ? "s" : ""} across all VMs
          </p>
        </div>
        <p className="text-xs text-text-muted">
          Checked {lastRefresh.toLocaleTimeString()}
        </p>
      </div>

      {error && (
        <div className="flex items-center gap-3 bg-surface border border-danger/30 rounded-lg px-4 py-3 text-sm text-text-primary">
          <span className="h-2 w-2 rounded-full bg-danger shrink-0" />
          Cannot reach control plane. Check connection.
        </div>
      )}

      {apps.length === 0 && !error ? (
        <div className="rounded-lg border border-border p-12 text-center">
          <p className="text-text-muted">No apps registered on this VM yet.</p>
          <p className="mt-1 text-sm text-text-muted">
            No VMs connected. Install the agent to get started.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {apps.map((app) => (
            <AppCard key={app.id} app={app} uptimePct={uptimes[app.id]} />
          ))}
        </div>
      )}
    </div>
  );
}
