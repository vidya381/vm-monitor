import Link from "next/link";
import { AppStatusBadge } from "@/components/status-badge";
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

export function AppCard({ app }: { app: App }) {
  return (
    <Link href={`/apps/${app.id}`}>
      <div className="bg-surface border border-border rounded-lg p-4 hover:border-border-subtle hover:bg-surface-raised transition-colors cursor-pointer">
        <div className="flex items-start justify-between gap-2 mb-1">
          <p className="text-base font-semibold text-text-primary">{app.name}</p>
          <AppStatusBadge status={app.last_status} />
        </div>
        <p className="text-sm text-text-secondary mb-4">{app.vm_name}</p>
        <div className="flex items-center justify-between">
          <span className="text-xs text-text-muted">{app.type}</span>
          <span className="text-xs text-text-muted">checked {timeAgo(app.last_checked_at)}</span>
        </div>
      </div>
    </Link>
  );
}
