import { NextResponse } from "next/server";
import { serverFetch } from "@/lib/server-api";

export async function GET(
  _req: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  try {
    const data = await serverFetch(`/apps/${id}/env`);
    return NextResponse.json(data);
  } catch (e) {
    return NextResponse.json({ error: String(e) }, { status: 502 });
  }
}

export async function PUT(
  req: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  const url = new URL(req.url);
  const restart = url.searchParams.get("restart") ?? "false";
  const body = await req.text();

  try {
    const data = await serverFetch(`/apps/${id}/env?restart=${restart}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body,
    });
    return NextResponse.json(data);
  } catch (e) {
    return NextResponse.json({ error: String(e) }, { status: 502 });
  }
}
