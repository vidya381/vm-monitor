import { NextResponse } from "next/server";
import { serverFetch } from "@/lib/server-api";
import { isDemoMode, demoEnvVars, DEMO_BLOCKED } from "@/lib/demo-data";

export async function GET(
  req: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;

  if (isDemoMode()) {
    return NextResponse.json(demoEnvVars[id] ?? {});
  }

  const url = new URL(req.url);
  const file = url.searchParams.get("file");
  const query = file ? `?file=${encodeURIComponent(file)}` : "";
  try {
    const data = await serverFetch(`/apps/${id}/env${query}`);
    return NextResponse.json(data);
  } catch (e) {
    return NextResponse.json({ error: String(e) }, { status: 502 });
  }
}

export async function PUT(
  req: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  if (isDemoMode()) return NextResponse.json(DEMO_BLOCKED, { status: 403 });

  const { id } = await params;
  const url = new URL(req.url);
  const restart = url.searchParams.get("restart") ?? "false";
  const file = url.searchParams.get("file");
  const body = await req.text();

  const qs = new URLSearchParams({ restart });
  if (file) qs.set("file", file);

  try {
    const data = await serverFetch(`/apps/${id}/env?${qs}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body,
    });
    return NextResponse.json(data);
  } catch (e) {
    return NextResponse.json({ error: String(e) }, { status: 502 });
  }
}
