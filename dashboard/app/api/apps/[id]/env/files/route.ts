import { NextResponse } from "next/server";
import { serverFetch } from "@/lib/server-api";

export async function GET(
  _req: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;
  try {
    const data = await serverFetch(`/apps/${id}/env/files`);
    return NextResponse.json(data);
  } catch (e) {
    return NextResponse.json([], { status: 200 });
  }
}
