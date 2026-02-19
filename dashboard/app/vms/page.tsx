"use client";

import { useState, useEffect } from "react";
import { VMStatusBadge } from "@/components/status-badge";
import { VM } from "@/lib/types";

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

export default function VMsPage() {
  const [vms, setVMs] = useState<VM[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch("/api/vms")
      .then((r) => r.json())
      .then((data) => { setVMs(data); setLoading(false); });
  }, []);

  if (loading) {
    return <div className="h-48 rounded-lg border border-border bg-muted animate-pulse" />;
  }

  return (
    <div className="space-y-4">
      <h1 className="text-xl font-semibold">Virtual Machines</h1>

      {vms.length === 0 ? (
        <div className="rounded-lg border border-dashed border-border p-12 text-center">
          <p className="text-muted-foreground">No VMs registered yet.</p>
        </div>
      ) : (
        <div className="rounded-lg border border-border divide-y divide-border">
          {vms.map((vm) => (
            <div key={vm.id} className="flex items-center justify-between px-4 py-3">
              <div>
                <p className="font-medium text-sm">{vm.name}</p>
                <p className="text-xs text-muted-foreground">
                  heartbeat {timeAgo(vm.last_heartbeat)}
                  {vm.labels?.length > 0 && ` · ${vm.labels.join(", ")}`}
                </p>
              </div>
              <VMStatusBadge status={vm.status} />
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
