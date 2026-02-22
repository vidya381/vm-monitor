import { NextResponse } from "next/server";
import { serverFetch } from "@/lib/server-api";

export async function GET(_req: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const data = await serverFetch(`/vms/${id}`);
  return NextResponse.json(data);
}
