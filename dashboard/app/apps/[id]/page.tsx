"use client";

import { use, useState, useEffect, useCallback } from "react";
import { AppStatusBadge } from "@/components/status-badge";
import { LogViewer } from "@/components/log-viewer";
import { EnvTab } from "@/components/env-tab";
import { AuditTab } from "@/components/audit-tab";
import { App, Metrics } from "@/lib/types";

function timeAgo(iso?: string) {
  if (!iso) return "never";
  const diff = Date.now() - new Date(iso).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  return `${Math.floor(hrs / 24)}d ago`;
}

const tabs = ["Status", "Logs", "Environment", "Audit"] as const;
type Tab = (typeof tabs)[number];

export default function AppDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const [app, setApp] = useState<App | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [activeTab, setActiveTab] = useState<Tab>("Status");
  const [restarting, setRestarting] = useState(false);
  const [confirmRestart, setConfirmRestart] = useState(false);
  const [restartMsg, setRestartMsg] = useState<{ type: "success" | "error"; text: string } | null>(null);
  const [metrics, setMetrics] = useState<Metrics | null>(null);

  const fetchApp = useCallback(() => {
    fetch(`/api/apps/${id}`)
      .then((r) => {
        if (!r.ok) throw new Error(`status ${r.status}`);
        return r.json();
      })
      .then((data) => { setApp(data); setLoading(false); })
      .catch((e) => { setError(String(e)); setLoading(false); });
  }, [id]);

  useEffect(() => { fetchApp(); }, [fetchApp]);

  useEffect(() => {
    if (activeTab !== "Status") return;
    const load = () =>
      fetch(`/api/apps/${id}/metrics`)
        .then((r) => (r.ok ? r.json() : null))
        .then((data) => setMetrics(data))
        .catch(() => setMetrics(null));
    load();
    const interval = setInterval(load, 30000);
    return () => clearInterval(interval);
  }, [id, activeTab]);

  async function handleRestart() {
    setConfirmRestart(false);
    setRestarting(true);
    setRestartMsg(null);
    try {
      const res = await fetch(`/api/apps/${id}/restart`, { method: "POST" });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error ?? `status ${res.status}`);
      setRestartMsg({ type: "success", text: `${app?.name} restarted successfully.` });
      setTimeout(fetchApp, 2000);
    } catch {
      setRestartMsg({ type: "error", text: `Cannot restart ${app?.name}. Check agent connection.` });
    } finally {
      setRestarting(false);
      setTimeout(() => setRestartMsg(null), 4000);
    }
  }

  if (loading) {
    return <div className="h-48 rounded-lg border border-border bg-surface animate-pulse" />;
  }

  if (error || !app) {
    return (
      <div className="flex items-center gap-3 bg-surface border border-danger/30 rounded-lg px-4 py-3 text-sm text-text-primary">
        <span className="h-2 w-2 rounded-full bg-danger shrink-0" />
        {error || "App not found"}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-xl font-semibold text-text-primary">{app.name}</h1>
          <p className="text-sm text-text-secondary">{app.vm_name}</p>
        </div>
        <AppStatusBadge status={app.last_status} />
      </div>

      {/* Tabs */}
      <div className="border-b border-border">
        <div className="flex gap-6">
          {tabs.map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={
                activeTab === tab
                  ? "pb-3 text-sm border-b-2 border-accent text-text-primary -mb-px"
                  : "pb-3 text-sm text-text-muted hover:text-text-secondary transition-colors"
              }
            >
              {tab}
            </button>
          ))}
        </div>
      </div>

      <div>
        {activeTab === "Status" && (
          <div className="space-y-4">
            {/* Restart toast */}
            {restartMsg && (
              <div
                className={`flex items-center gap-3 rounded-lg px-4 py-3 text-sm text-text-primary border ${
                  restartMsg.type === "success"
                    ? "bg-surface border-status-running/30"
                    : "bg-surface border-danger/30"
                }`}
              >
                <span
                  className={`h-2 w-2 rounded-full shrink-0 ${
                    restartMsg.type === "success" ? "bg-status-running" : "bg-danger"
                  }`}
                />
                {restartMsg.text}
              </div>
            )}

            <div className="rounded-lg border border-border divide-y divide-border">
              <Row label="App name" value={app.name} />
              <Row label="VM" value={app.vm_name} />
              <Row label="Type" value={app.type} />
              <Row label="Environment" value={app.environment || "—"} />
              <Row label="Status" value={<AppStatusBadge status={app.last_status} />} />
              <Row label="Last checked" value={timeAgo(app.last_checked_at)} />
              <Row label="Last restarted" value={timeAgo(app.last_restarted_at)} />
              <Row label="CPU" value={metrics ? `${metrics.cpu_percent}%` : "—"} />
              <Row label="Memory" value={metrics ? `${metrics.mem_rss_mb} MB` : "—"} />
              <Row label="PID" value={metrics ? String(metrics.pid) : "—"} />
            </div>

            {/* Restart action */}
            <div className="flex items-center gap-3">
              {confirmRestart ? (
                <>
                  <span className="text-sm text-text-secondary">
                    Restart {app.name}? This will briefly interrupt the service.
                  </span>
                  <button
                    onClick={handleRestart}
                    disabled={restarting}
                    className="bg-transparent hover:bg-danger-subtle text-danger border border-danger/30 hover:border-danger/60 text-sm font-medium px-4 py-2 rounded-md transition-colors disabled:opacity-50"
                  >
                    Confirm
                  </button>
                  <button
                    onClick={() => setConfirmRestart(false)}
                    className="bg-transparent hover:bg-surface-raised text-text-secondary hover:text-text-primary border border-border text-sm font-medium px-4 py-2 rounded-md transition-colors"
                  >
                    Cancel
                  </button>
                </>
              ) : (
                <button
                  onClick={() => setConfirmRestart(true)}
                  disabled={restarting}
                  className="bg-transparent hover:bg-danger-subtle text-danger border border-danger/30 hover:border-danger/60 text-sm font-medium px-4 py-2 rounded-md transition-colors disabled:opacity-50"
                >
                  {restarting ? "Restarting..." : "Restart App"}
                </button>
              )}
            </div>
          </div>
        )}

        {activeTab === "Logs" && <LogViewer appId={id} />}
        {activeTab === "Environment" && <EnvTab appId={id} />}
        {activeTab === "Audit" && <AuditTab appId={id} />}
      </div>
    </div>
  );
}

function Row({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between px-4 py-3 text-sm">
      <span className="text-text-muted">{label}</span>
      <span className="text-text-primary font-medium">{value}</span>
    </div>
  );
}
