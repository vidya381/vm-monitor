import { Badge } from "@/components/ui/badge";
import { AppStatus, VMStatus } from "@/lib/types";

const appStatusConfig: Record<string, { label: string; className: string }> = {
  running:   { label: "Running",   className: "bg-green-100 text-green-800 border-green-200" },
  stopped:   { label: "Stopped",   className: "bg-zinc-100 text-zinc-600 border-zinc-200" },
  unhealthy: { label: "Unhealthy", className: "bg-red-100 text-red-800 border-red-200" },
};

const vmStatusConfig: Record<string, { label: string; className: string }> = {
  online:      { label: "Online",      className: "bg-green-100 text-green-800 border-green-200" },
  unreachable: { label: "Unreachable", className: "bg-red-100 text-red-800 border-red-200" },
  unknown:     { label: "Unknown",     className: "bg-zinc-100 text-zinc-500 border-zinc-200" },
};

export function AppStatusBadge({ status }: { status: AppStatus }) {
  const cfg = appStatusConfig[status] ?? { label: status || "Unknown", className: "bg-zinc-100 text-zinc-500 border-zinc-200" };
  return (
    <Badge variant="outline" className={cfg.className}>
      <span className="mr-1.5 inline-block h-1.5 w-1.5 rounded-full bg-current" />
      {cfg.label}
    </Badge>
  );
}

export function VMStatusBadge({ status }: { status: VMStatus }) {
  const cfg = vmStatusConfig[status] ?? vmStatusConfig.unknown;
  return (
    <Badge variant="outline" className={cfg.className}>
      {cfg.label}
    </Badge>
  );
}
