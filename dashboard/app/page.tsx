"use client";

import { useState, useEffect } from "react";
import { AppCard } from "@/components/app-card";
import { App } from "@/lib/types";

export default function OverviewPage() {
  const [apps, setApps] = useState<App[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [lastRefresh, setLastRefresh] = useState(new Date());

  const fetchApps = async () => {
    try {
      const res = await fetch("/api/apps");
      if (!res.ok) throw new Error(`status ${res.status}`);
      setApps(await res.json());
      setLastRefresh(new Date());
      setError("");
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
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {[...Array(3)].map((_, i) => (
          <div key={i} className="h-32 rounded-lg border border-border bg-muted animate-pulse" />
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold">Apps</h1>
          <p className="text-sm text-muted-foreground">
            {apps.length} app{apps.length !== 1 ? "s" : ""} across all VMs
          </p>
        </div>
        <p className="text-xs text-muted-foreground">
          refreshed {lastRefresh.toLocaleTimeString()}
        </p>
      </div>

      {error && (
        <div className="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
          Could not reach control plane: {error}
        </div>
      )}

      {apps.length === 0 && !error ? (
        <div className="rounded-lg border border-dashed border-border p-12 text-center">
          <p className="text-muted-foreground">No apps registered yet.</p>
          <p className="mt-1 text-sm text-muted-foreground">
            Install the agent on a VM to get started.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {apps.map((app) => (
            <AppCard key={app.id} app={app} />
          ))}
        </div>
      )}
    </div>
  );
}
