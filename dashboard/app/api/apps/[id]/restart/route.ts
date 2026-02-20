import { NextResponse } from "next/server";
import { serverFetch } from "@/lib/server-api";

export async function POST(
  _req: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  try {
    const data = await serverFetch(`/apps/${id}/restart`, { method: "POST" });
    return NextResponse.json(data);
  } catch (e) {
    return NextResponse.json({ error: String(e) }, { status: 502 });
  }
}
