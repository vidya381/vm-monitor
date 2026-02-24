import { NextResponse } from "next/server";
import { serverFetch } from "@/lib/server-api";
import { isDemoMode, demoMetrics } from "@/lib/demo-data";

export async function GET(
  _req: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;

  if (isDemoMode()) {
    const m = demoMetrics[id];
    if (!m) return NextResponse.json(null, { status: 503 });
    return NextResponse.json(m);
  }

  try {
    const data = await serverFetch(`/apps/${id}/metrics`);
    return NextResponse.json(data);
  } catch {
    // App may be stopped (no PID) — return null so the UI shows "not available"
    return NextResponse.json(null, { status: 503 });
  }
}
