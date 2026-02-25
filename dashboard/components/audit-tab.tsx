"use client";

import { useState, useEffect } from "react";
import { AuditLog } from "@/lib/types";

function timeAgo(iso: string) {
  const diff = Date.now() - new Date(iso).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  return `${Math.floor(hrs / 24)}d ago`;
}

const actionStyle: Record<string, string> = {
  env_update: "text-accent border border-accent-muted bg-accent-subtle",
  restart:    "text-warning border border-warning/30 bg-status-unhealthy-bg",
};

function ActionBadge({ action }: { action: string }) {
  const cls = actionStyle[action] ?? "text-text-muted border border-border bg-surface";
  return (
    <span className={`font-mono text-xs px-2 py-0.5 rounded-md ${cls}`}>
      {action}
    </span>
  );
}

function formatDetails(details: Record<string, unknown> | null): string {
  if (!details) return "";
  const parts: string[] = [];
  if (typeof details.keys_updated === "number") {
    parts.push(`${details.keys_updated} key${details.keys_updated !== 1 ? "s" : ""} updated`);
  }
  if (details.restart === true) parts.push("restarted");
  return parts.join(" · ");
}

export function AuditTab({ appId }: { appId: string }) {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    fetch(`/api/apps/${appId}/audit`)
      .then((r) => r.json())
      .then((data) => { setLogs(data); setLoading(false); })
      .catch((e) => { setError(String(e)); setLoading(false); });
  }, [appId]);

  if (loading) {
    return <div className="h-48 rounded-lg border border-border bg-surface animate-pulse" />;
  }

  if (error) {
    return (
      <div className="flex items-center gap-3 bg-surface border border-danger/30 rounded-lg px-4 py-3 text-sm text-text-primary">
        <span className="h-2 w-2 rounded-full bg-danger shrink-0" />
        Cannot load audit log. Check agent connection.
      </div>
    );
  }

  if (logs.length === 0) {
    return (
      <div className="rounded-lg border border-border p-12 text-center">
        <p className="text-text-muted">No changes recorded yet.</p>
      </div>
    );
  }

  return (
    <div className="rounded-lg border border-border divide-y divide-border">
      {logs.map((log) => (
        <div key={log.id} className="flex flex-wrap items-center gap-x-3 gap-y-1 px-4 py-3 hover:bg-surface-raised transition-colors">
          <span className="text-xs text-text-muted shrink-0 min-w-[52px]" title={log.created_at}>
            {timeAgo(log.created_at)}
          </span>
          <ActionBadge action={log.action} />
          <span className="text-xs text-text-muted">
            {formatDetails(log.details)}
          </span>
        </div>
      ))}
    </div>
  );
}
