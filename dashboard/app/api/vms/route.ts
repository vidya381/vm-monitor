import { NextResponse } from "next/server";
import { serverFetch } from "@/lib/server-api";
import { isDemoMode, demoVMs } from "@/lib/demo-data";

export async function GET() {
  if (isDemoMode()) return NextResponse.json(demoVMs);
  const data = await serverFetch("/vms");
  return NextResponse.json(data);
}
