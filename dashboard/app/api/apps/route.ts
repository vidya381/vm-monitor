import { NextResponse } from "next/server";
import { serverFetch } from "@/lib/server-api";
import { isDemoMode, demoApps, DEMO_BLOCKED } from "@/lib/demo-data";

export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const vmId = searchParams.get("vm_id");

  if (isDemoMode()) {
    const apps = vmId ? demoApps.filter((a) => a.vm_id === vmId) : demoApps;
    return NextResponse.json(apps);
  }

  const path = vmId ? `/apps?vm_id=${vmId}` : "/apps";
  const data = await serverFetch(path);
  return NextResponse.json(data);
}

export async function POST(req: Request) {
  if (isDemoMode()) return NextResponse.json(DEMO_BLOCKED, { status: 403 });
  const body = await req.json();
  const data = await serverFetch("/apps", {
    method: "POST",
    body: JSON.stringify(body),
  });
  return NextResponse.json(data, { status: 201 });
}
