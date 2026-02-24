import { NextResponse } from "next/server";
import { serverFetch } from "@/lib/server-api";
import { isDemoMode, demoVMs } from "@/lib/demo-data";

export async function GET(_req: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  if (isDemoMode()) {
    const vm = demoVMs.find((v) => v.id === id);
    if (!vm) return NextResponse.json({ error: "not found" }, { status: 404 });
    return NextResponse.json(vm);
  }

  const data = await serverFetch(`/vms/${id}`);
  return NextResponse.json(data);
}
