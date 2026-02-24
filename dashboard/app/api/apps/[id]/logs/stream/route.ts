// SSE passthrough — cannot use serverFetch (it buffers via .json()).
// Instead we pipe the upstream response body directly to the client.
import { isDemoMode } from "@/lib/demo-data";

const API_URL = process.env.API_URL!;
const API_KEY = process.env.API_KEY!;

export async function GET(
  _req: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params;

  // In demo mode, return a non-2xx so the LogViewer falls back to HTTP polling.
  if (isDemoMode()) {
    return new Response(null, { status: 204 });
  }

  const upstream = await fetch(`${API_URL}/apps/${id}/logs/stream`, {
    headers: { Authorization: `Bearer ${API_KEY}` },
  });

  if (!upstream.ok || !upstream.body) {
    return new Response("upstream error", { status: upstream.status ?? 502 });
  }

  return new Response(upstream.body, {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      "Connection": "keep-alive",
    },
  });
}
