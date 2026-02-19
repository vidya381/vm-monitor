import { NextResponse } from "next/server";
import { serverFetch } from "@/lib/server-api";

export async function GET() {
  const data = await serverFetch("/apps");
  return NextResponse.json(data);
}
