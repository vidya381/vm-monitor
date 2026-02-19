"use client";

import { use, useState, useEffect } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
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

export default function AppDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const [app, setApp] = useState<App | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

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
    return <div className="h-48 rounded-lg border border-border bg-muted animate-pulse" />;
  }

  if (error || !app) {
    return (
      <div className="rounded-md border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">
        {error || "App not found"}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-xl font-semibold">{app.name}</h1>
          <p className="text-sm text-muted-foreground">{app.vm_name}</p>
        </div>
        <AppStatusBadge status={app.last_status} />
      </div>

      <Tabs defaultValue="status">
        <TabsList>
          <TabsTrigger value="status">Status</TabsTrigger>
          <TabsTrigger value="logs">Logs</TabsTrigger>
        </TabsList>

        {/* Status tab */}
        <TabsContent value="status" className="mt-4">
          <div className="rounded-lg border border-border divide-y divide-border">
            <Row label="App name" value={app.name} />
            <Row label="VM" value={app.vm_name} />
            <Row label="Type" value={app.type} />
            <Row label="Environment" value={app.environment || "—"} />
            <Row label="Status" value={<AppStatusBadge status={app.last_status} />} />
            <Row label="Last checked" value={timeAgo(app.last_checked_at)} />
            <Row label="Last restarted" value={timeAgo(app.last_restarted_at)} />
          </div>
        </TabsContent>

        {/* Logs tab */}
        <TabsContent value="logs" className="mt-4">
          <LogViewer appId={id} />
        </TabsContent>
      </Tabs>
    </div>
  );
}

function Row({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between px-4 py-3 text-sm">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-medium">{value}</span>
    </div>
  );
}
