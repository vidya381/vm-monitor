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
    const data = await serverFetch(`/apps/${id}/deploy`, { method: "POST" });
    return NextResponse.json(data);
  } catch {
    return NextResponse.json(
      { success: false, output: "", error: "Agent unreachable" },
      { status: 503 }
    );
  }
}
