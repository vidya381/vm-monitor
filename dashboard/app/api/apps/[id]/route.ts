import { NextResponse } from "next/server";
import { serverFetch } from "@/lib/server-api";
import { isDemoMode, demoApps } from "@/lib/demo-data";

export async function GET(
  _req: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;

  if (isDemoMode()) {
    const app = demoApps.find((a) => a.id === id);
    if (!app) return NextResponse.json({ error: "not found" }, { status: 404 });
    return NextResponse.json(app);
  }

  const data = await serverFetch(`/apps/${id}`);
  return NextResponse.json(data);
}
