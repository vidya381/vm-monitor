import { NextResponse } from "next/server";
import { serverFetch } from "@/lib/server-api";
import { isDemoMode, demoAudit } from "@/lib/demo-data";

export async function GET(
  _req: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;

  if (isDemoMode()) {
    return NextResponse.json(demoAudit[id] ?? []);
  }

  const data = await serverFetch(`/apps/${id}/audit`);
  return NextResponse.json(data);
}
