"use client";

import { use, useState, useEffect } from "react";
import Link from "next/link";
import { ChevronLeft } from "lucide-react";
import { VMStatusBadge } from "@/components/status-badge";
import { AppCard } from "@/components/app-card";
import { VM, App } from "@/lib/types";

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

export default function VMDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const [vm, setVM] = useState<VM | null>(null);
  const [apps, setApps] = useState<App[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    Promise.all([
      fetch(`/api/vms/${id}`).then((r) => {
        if (!r.ok) throw new Error(`status ${r.status}`);
        return r.json();
      }),
      fetch(`/api/apps?vm_id=${id}`).then((r) => {
        if (!r.ok) throw new Error(`status ${r.status}`);
        return r.json();
      }),
    ])
      .then(([vmData, appsData]) => {
        setVM(vmData);
        setApps(appsData);
        setLoading(false);
      })
      .catch((e) => {
        setError(String(e));
        setLoading(false);
      });
  }, [id]);

  if (loading) {
    return <div className="h-48 rounded-lg border border-border bg-surface animate-pulse" />;
  }

  if (error || !vm) {
    return (
      <div className="flex items-center gap-3 bg-surface border border-danger/30 rounded-lg px-4 py-3 text-sm text-text-primary">
        <span className="h-2 w-2 rounded-full bg-danger shrink-0" />
        {error || "VM not found"}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Back link */}
      <Link
        href="/vms"
        className="inline-flex items-center gap-1 text-sm text-text-muted hover:text-text-primary transition-colors"
      >
        <ChevronLeft size={14} />
        VMs
      </Link>

      {/* Header */}
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-xl font-semibold text-text-primary">{vm.name}</h1>
          <p className="text-sm text-text-muted">
            heartbeat {timeAgo(vm.last_heartbeat)}
            {vm.labels?.length > 0 && ` · ${vm.labels.join(", ")}`}
          </p>
        </div>
        <VMStatusBadge status={vm.status} />
      </div>

      {/* Apps */}
      <div>
        <p className="text-sm text-text-muted mb-3">
          {apps.length} {apps.length === 1 ? "app" : "apps"}
        </p>
        {apps.length === 0 ? (
          <div className="rounded-lg border border-border p-12 text-center">
            <p className="text-text-muted text-sm">No apps registered for this VM.</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {apps.map((app) => (
              <AppCard key={app.id} app={app} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
