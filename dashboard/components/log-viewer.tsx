"use client";

import { useState, useEffect, useRef, useCallback } from "react";

interface LogViewerProps {
  appId: string;
}

export function LogViewer({ appId }: LogViewerProps) {
  const [lines, setLines] = useState<string[]>([]);
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const cursorRef = useRef<string>("");
  const bottomRef = useRef<HTMLDivElement>(null);

  const fetchLogs = useCallback(async (initial = false) => {
    try {
      const qs = new URLSearchParams();
      if (initial) {
        qs.set("tail", "200");
      } else if (cursorRef.current) {
        qs.set("cursor", cursorRef.current);
      } else {
        return;
      }

      const res = await fetch(`/api/apps/${appId}/logs?${qs.toString()}`);
      if (!res.ok) throw new Error(`status ${res.status}`);
      const data = await res.json();

      if (data.lines?.length > 0) {
        setLines((prev) => (initial ? data.lines : [...prev, ...data.lines]));
        if (data.cursor) cursorRef.current = data.cursor;
      }
      if (initial) setLoading(false);
    } catch (e) {
      setError(String(e));
      if (initial) setLoading(false);
    }
  }, [appId]);

  // Auto-scroll to bottom when new lines arrive
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [lines]);

  useEffect(() => {
    fetchLogs(true);
    const interval = setInterval(() => fetchLogs(false), 5000);
    return () => clearInterval(interval);
  }, [fetchLogs]);

  const filtered = search
    ? lines.filter((l) => l.toLowerCase().includes(search.toLowerCase()))
    : lines;

  return (
    <div className="flex flex-col gap-3">
      <input
        type="text"
        placeholder="Search logs…"
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        className="w-full rounded-md border border-input bg-background px-3 py-1.5 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring"
      />

      <div className="h-[480px] overflow-y-auto rounded-md border border-border bg-zinc-950 p-4 font-mono text-xs text-zinc-100">
        {loading && (
          <p className="text-zinc-400">Loading logs…</p>
        )}
        {error && (
          <p className="text-red-400">Error: {error}</p>
        )}
        {!loading && filtered.length === 0 && (
          <p className="text-zinc-500">{search ? "No matching lines." : "No logs yet."}</p>
        )}
        {filtered.map((line, i) => (
          <div key={i} className="leading-5 whitespace-pre-wrap break-all">
            {line}
          </div>
        ))}
        <div ref={bottomRef} />
      </div>

      <p className="text-xs text-muted-foreground">
        {lines.length} lines · polls every 5s
        {search && ` · ${filtered.length} matching`}
      </p>
    </div>
  );
}
