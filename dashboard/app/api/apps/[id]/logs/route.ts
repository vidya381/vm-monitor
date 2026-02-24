import { NextResponse } from "next/server";
import { serverFetch } from "@/lib/server-api";
import { isDemoMode, demoLogs } from "@/lib/demo-data";

export async function GET(
  req: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;

  if (isDemoMode()) {
    return NextResponse.json({ lines: demoLogs[id] ?? [] });
  }

  const { searchParams } = new URL(req.url);

  const qs = new URLSearchParams();
  const tail = searchParams.get("tail");
  const cursor = searchParams.get("cursor");
  if (tail) qs.set("tail", tail);
  if (cursor) qs.set("cursor", cursor);

  const data = await serverFetch(`/apps/${id}/logs?${qs.toString()}`);
  return NextResponse.json(data);
}
