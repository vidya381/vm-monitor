import { NextResponse } from "next/server";
import { serverFetch } from "@/lib/server-api";
import { isDemoMode, demoUptime } from "@/lib/demo-data";

export async function GET(
  req: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const { searchParams } = new URL(req.url);
  const days = searchParams.get("days") ?? "30";

  if (isDemoMode()) {
    const u = demoUptime[id];
    if (!u) return NextResponse.json({ uptime_pct: null, incident_count: 0, incidents: [] });
    return NextResponse.json(u);
  }

  try {
    const data = await serverFetch(`/apps/${id}/uptime?days=${days}`);
    return NextResponse.json(data);
  } catch {
    return NextResponse.json(null, { status: 503 });
  }
}
