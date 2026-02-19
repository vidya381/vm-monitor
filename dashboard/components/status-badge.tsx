import { AppStatus, VMStatus } from "@/lib/types";

const appStatusConfig: Record<string, { dot: string; label: string; pulse: boolean }> = {
  running:   { dot: "bg-status-running",   label: "Running",   pulse: true  },
  stopped:   { dot: "bg-status-stopped",   label: "Stopped",   pulse: false },
  unhealthy: { dot: "bg-status-unhealthy", label: "Unhealthy", pulse: false },
};

const vmStatusConfig: Record<string, { dot: string; label: string }> = {
  online:      { dot: "bg-status-running",  label: "Online"      },
  unreachable: { dot: "bg-status-stopped",  label: "Unreachable" },
  unknown:     { dot: "bg-status-unknown",  label: "Unknown"     },
};

export function AppStatusBadge({ status }: { status: AppStatus }) {
  const cfg = appStatusConfig[status] ?? { dot: "bg-status-unknown", label: status || "Unknown", pulse: false };
  return (
    <span className="flex items-center gap-1.5">
      <span className={`h-2 w-2 rounded-full ${cfg.dot}${cfg.pulse ? " animate-pulse" : ""}`} />
      <span className="text-xs text-text-muted">{cfg.label}</span>
    </span>
  );
}

export function VMStatusBadge({ status }: { status: VMStatus }) {
  const cfg = vmStatusConfig[status] ?? vmStatusConfig.unknown;
  return (
    <span className="flex items-center gap-1.5">
      <span className={`h-2 w-2 rounded-full ${cfg.dot}`} />
      <span className="text-xs text-text-muted">{cfg.label}</span>
    </span>
  );
}
