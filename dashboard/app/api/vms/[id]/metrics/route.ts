import { NextResponse } from "next/server";
import { serverFetch } from "@/lib/server-api";
import { isDemoMode, demoSystemMetrics } from "@/lib/demo-data";

export async function GET(
  _req: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;

  if (isDemoMode()) {
    const m = demoSystemMetrics[id];
    if (!m) return NextResponse.json(null, { status: 503 });
    return NextResponse.json(m);
  }

  try {
    const data = await serverFetch(`/vms/${id}/metrics`);
    return NextResponse.json(data);
  } catch {
    return NextResponse.json(null, { status: 503 });
  }
}
