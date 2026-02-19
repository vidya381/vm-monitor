"use client";

import { use, useState } from "react";
import { useEffect } from "react";
import { AppStatusBadge } from "@/components/status-badge";
import { LogViewer } from "@/components/log-viewer";
import { App } from "@/lib/types";

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

const tabs = ["Status", "Logs"] as const;
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

  useEffect(() => {
    fetch(`/api/apps/${id}`)
      .then((r) => {
        if (!r.ok) throw new Error(`status ${r.status}`);
        return r.json();
      })
      .then((data) => { setApp(data); setLoading(false); })
      .catch((e) => { setError(String(e)); setLoading(false); });
  }, [id]);

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

      <div className="pt-0">
        {activeTab === "Status" && (
          <div className="rounded-lg border border-border divide-y divide-border">
            <Row label="App name" value={app.name} />
            <Row label="VM" value={app.vm_name} />
            <Row label="Type" value={app.type} />
            <Row label="Environment" value={app.environment || "—"} />
            <Row label="Status" value={<AppStatusBadge status={app.last_status} />} />
            <Row label="Last checked" value={timeAgo(app.last_checked_at)} />
            <Row label="Last restarted" value={timeAgo(app.last_restarted_at)} />
          </div>
        )}
        {activeTab === "Logs" && <LogViewer appId={id} />}
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
