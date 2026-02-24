import { NextResponse } from "next/server";
import { serverFetch } from "@/lib/server-api";

export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const vmId = searchParams.get("vm_id");
  const path = vmId ? `/apps?vm_id=${vmId}` : "/apps";
  const data = await serverFetch(path);
  return NextResponse.json(data);
}

export async function POST(req: Request) {
  const body = await req.json();
  const data = await serverFetch("/apps", {
    method: "POST",
    body: JSON.stringify(body),
  });
  return NextResponse.json(data, { status: 201 });
}
