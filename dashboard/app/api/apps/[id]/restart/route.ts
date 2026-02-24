import { NextResponse } from "next/server";
import { serverFetch } from "@/lib/server-api";
import { isDemoMode, DEMO_BLOCKED } from "@/lib/demo-data";

export async function POST(
  _req: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  if (isDemoMode()) return NextResponse.json(DEMO_BLOCKED, { status: 403 });

  const { id } = await params;
  try {
    const data = await serverFetch(`/apps/${id}/restart`, { method: "POST" });
    return NextResponse.json(data);
  } catch (e) {
    return NextResponse.json({ error: String(e) }, { status: 502 });
  }
}
