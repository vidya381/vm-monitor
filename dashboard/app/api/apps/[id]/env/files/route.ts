import { NextResponse } from "next/server";
import { serverFetch } from "@/lib/server-api";
import { isDemoMode, demoEnvFiles } from "@/lib/demo-data";

export async function GET(
  _req: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;

  if (isDemoMode()) {
    return NextResponse.json(demoEnvFiles[id] ?? []);
  }

  try {
    const data = await serverFetch(`/apps/${id}/env/files`);
    return NextResponse.json(data);
  } catch (e) {
    return NextResponse.json([], { status: 200 });
  }
}
