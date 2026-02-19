import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
      <Card className="hover:shadow-md transition-shadow cursor-pointer h-full">
        <CardHeader className="pb-2">
          <div className="flex items-start justify-between gap-2">
            <CardTitle className="text-base font-semibold">{app.name}</CardTitle>
            <AppStatusBadge status={app.last_status} />
          </div>
          <p className="text-sm text-muted-foreground">{app.vm_name}</p>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between text-xs text-muted-foreground">
            <span className="uppercase tracking-wide font-medium">{app.type}</span>
            <span>checked {timeAgo(app.last_checked_at)}</span>
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}
