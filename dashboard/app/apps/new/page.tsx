"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { ChevronLeft } from "lucide-react";
import { VM } from "@/lib/types";

export default function NewAppPage() {
  const router = useRouter();
  const [vms, setVMs] = useState<VM[]>([]);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  const [vmId, setVmId] = useState("");
  const [name, setName] = useState("");
  const [appType, setAppType] = useState<"systemd" | "docker">("systemd");
  const [service, setService] = useState("");
  const [container, setContainer] = useState("");
  const [deployDir, setDeployDir] = useState("");
  const [environment, setEnvironment] = useState("");
  const [autoRestart, setAutoRestart] = useState(false);

  useEffect(() => {
    fetch("/api/vms")
      .then((r) => r.json())
      .then(setVMs)
      .catch(() => {});
  }, []);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");

    if (!vmId) { setError("Please select a VM."); return; }
    if (!name.trim()) { setError("App name is required."); return; }
    if (appType === "systemd" && !service.trim()) { setError("Service name is required for systemd apps."); return; }
    if (appType === "docker" && !container.trim()) { setError("Container name is required for docker apps."); return; }

    setSubmitting(true);
    try {
      const res = await fetch("/api/apps", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          vm_id: vmId,
          name: name.trim(),
          type: appType,
          environment: environment.trim(),
          config: {
            service: appType === "systemd" ? service.trim() : undefined,
            container: appType === "docker" ? container.trim() : undefined,
            deploy_dir: deployDir.trim() || undefined,
            auto_restart: autoRestart,
          },
        }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error ?? `status ${res.status}`);
      router.push(`/apps/${data.id}`);
    } catch (err) {
      setError(String(err));
      setSubmitting(false);
    }
  }

  const inputClass =
    "w-full bg-background border border-border hover:border-border-subtle focus:border-accent rounded-md px-3 py-2 text-sm text-text-primary outline-none transition-colors";
  const labelClass = "block text-sm text-text-muted mb-1";

  return (
    <div className="space-y-6 max-w-lg">
      <Link
        href="/"
        className="inline-flex items-center gap-1 text-sm text-text-muted hover:text-text-primary transition-colors"
      >
        <ChevronLeft size={14} />
        Apps
      </Link>

      <div>
        <h1 className="text-xl font-semibold text-text-primary">Register App</h1>
        <p className="text-sm text-text-muted mt-1">
          Add an existing app to the dashboard. Set up the service on the VM first, then register it here.
        </p>
      </div>

      {error && (
        <div className="flex items-center gap-3 bg-surface border border-danger/30 rounded-lg px-4 py-3 text-sm text-text-primary">
          <span className="h-2 w-2 rounded-full bg-danger shrink-0" />
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        {/* VM */}
        <div>
          <label className={labelClass}>VM <span className="text-danger">*</span></label>
          <select
            value={vmId}
            onChange={(e) => setVmId(e.target.value)}
            className={inputClass}
          >
            <option value="">Select a VM…</option>
            {vms.map((vm) => (
              <option key={vm.id} value={vm.id}>
                {vm.name}
              </option>
            ))}
          </select>
        </div>

        {/* App name */}
        <div>
          <label className={labelClass}>App Name <span className="text-danger">*</span></label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="my-app"
            className={inputClass}
          />
        </div>

        {/* Type */}
        <div>
          <label className={labelClass}>Type <span className="text-danger">*</span></label>
          <div className="flex gap-3">
            {(["systemd", "docker"] as const).map((t) => (
              <button
                key={t}
                type="button"
                onClick={() => setAppType(t)}
                className={`flex-1 py-2 rounded-md border text-sm font-medium transition-colors ${
                  appType === t
                    ? "border-accent text-accent bg-accent/5"
                    : "border-border text-text-muted hover:text-text-primary hover:border-border-subtle"
                }`}
              >
                {t}
              </button>
            ))}
          </div>
        </div>

        {/* Service / Container */}
        {appType === "systemd" ? (
          <div>
            <label className={labelClass}>Service Name <span className="text-danger">*</span></label>
            <input
              type="text"
              value={service}
              onChange={(e) => setService(e.target.value)}
              placeholder="myapp.service"
              className={inputClass}
            />
          </div>
        ) : (
          <div>
            <label className={labelClass}>Container Name <span className="text-danger">*</span></label>
            <input
              type="text"
              value={container}
              onChange={(e) => setContainer(e.target.value)}
              placeholder="myapp"
              className={inputClass}
            />
          </div>
        )}

        {/* Deploy dir */}
        <div>
          <label className={labelClass}>Working Directory (for git pull)</label>
          <input
            type="text"
            value={deployDir}
            onChange={(e) => setDeployDir(e.target.value)}
            placeholder="/home/ubuntu/myapp"
            className={inputClass}
          />
          <p className="text-xs text-text-muted mt-1">Leave empty if you don&apos;t need git deploy from the dashboard.</p>
        </div>

        {/* Environment */}
        <div>
          <label className={labelClass}>Environment</label>
          <input
            type="text"
            value={environment}
            onChange={(e) => setEnvironment(e.target.value)}
            placeholder="production"
            className={inputClass}
          />
        </div>

        {/* Auto restart */}
        <label className="flex items-center gap-3 cursor-pointer select-none">
          <div
            onClick={() => setAutoRestart(!autoRestart)}
            className={`w-9 h-5 rounded-full transition-colors relative ${autoRestart ? "bg-accent" : "bg-border"}`}
          >
            <span
              className={`absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full transition-transform ${autoRestart ? "translate-x-4" : ""}`}
            />
          </div>
          <span className="text-sm text-text-secondary">Auto-restart on crash</span>
        </label>

        {/* Submit */}
        <div className="flex gap-3 pt-2">
          <button
            type="submit"
            disabled={submitting}
            className="bg-accent hover:bg-accent-hover text-background text-sm font-medium px-5 py-2 rounded-md transition-colors disabled:opacity-50"
          >
            {submitting ? "Registering…" : "Register App"}
          </button>
          <Link
            href="/"
            className="bg-transparent hover:bg-surface-raised text-text-secondary hover:text-text-primary border border-border text-sm font-medium px-5 py-2 rounded-md transition-colors"
          >
            Cancel
          </Link>
        </div>
      </form>
    </div>
  );
}
