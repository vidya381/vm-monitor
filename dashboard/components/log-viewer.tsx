"use client";

import { useState, useEffect, useRef, useCallback } from "react";

export function LogViewer({ appId }: { appId: string }) {
  const [lines, setLines] = useState<string[]>([]);
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [paused, setPaused] = useState(false);
  const [autoScroll, setAutoScroll] = useState(true);
  const [streaming, setStreaming] = useState(false);
  const cursorRef = useRef<string>("");
  const bottomRef = useRef<HTMLDivElement>(null);
  const pausedRef = useRef(false);
  pausedRef.current = paused;

  const fetchLogs = useCallback(async (initial = false) => {
    if (!initial && pausedRef.current) return;
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

  useEffect(() => {
    if (autoScroll) bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [lines, autoScroll]);

  useEffect(() => {
    // Always load initial 200 lines via HTTP first.
    fetchLogs(true);

    // Then try SSE for live updates.
    const es = new EventSource(`/api/apps/${appId}/logs/stream`);
    let connected = false;
    let pollInterval: ReturnType<typeof setInterval> | null = null;

    es.onopen = () => {
      connected = true;
      setStreaming(true);
    };

    es.onmessage = (e) => {
      if (pausedRef.current) return;
      try {
        const { line } = JSON.parse(e.data);
        if (line) setLines((prev) => [...prev, line]);
      } catch {}
    };

    es.onerror = () => {
      if (!connected) {
        // SSE failed to connect — fall back to cursor polling.
        es.close();
        setStreaming(false);
        pollInterval = setInterval(() => fetchLogs(false), 5000);
      }
    };

    return () => {
      es.close();
      if (pollInterval) clearInterval(pollInterval);
    };
  }, [appId, fetchLogs]);

  const filtered = search
    ? lines.filter((l) => l.toLowerCase().includes(search.toLowerCase()))
    : lines;

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center gap-3">
        <input
          type="text"
          placeholder="filter by content..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="flex-1 bg-background border border-border hover:border-border-subtle focus:border-accent rounded-md px-3 py-2 text-sm text-text-primary placeholder:text-text-disabled outline-none transition-colors font-mono"
        />
        <button
          onClick={() => setPaused((p) => !p)}
          className="flex items-center gap-2 bg-transparent hover:bg-surface-raised text-text-secondary hover:text-text-primary border border-border text-sm font-medium px-4 py-2 rounded-md transition-colors"
        >
          {paused && <span className="h-1.5 w-1.5 rounded-full bg-warning" />}
          {paused ? "Paused" : "Pause"}
        </button>
      </div>

      <div className="bg-background border border-border rounded-lg font-mono text-xs text-text-secondary h-64 sm:h-96 overflow-y-auto p-4 space-y-0.5">
        {loading && <p className="text-text-muted">Loading logs...</p>}
        {error && <p className="text-status-stopped">Cannot load logs. Check agent connection.</p>}
        {!loading && !error && filtered.length === 0 && (
          <p className="text-text-muted">{search ? "No matching lines." : "No logs available."}</p>
        )}
        {filtered.map((line, i) => (
          <div key={i} className="flex gap-4 hover:bg-surface-raised px-1 rounded">
            <span className="text-text-disabled w-8 shrink-0 select-none text-right">{i + 1}</span>
            <span className="break-all">{line}</span>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>

      <div className="flex items-center justify-between">
        <p className="text-xs text-text-muted">
          {lines.length} lines{search && ` · ${filtered.length} matching`}
          {streaming && <span className="ml-2 text-status-running">● live</span>}
        </p>
        <label className="flex items-center gap-2 text-xs text-text-muted cursor-pointer select-none">
          <input
            type="checkbox"
            checked={autoScroll}
            onChange={(e) => setAutoScroll(e.target.checked)}
            className="accent-accent"
          />
          Auto-scroll
        </label>
      </div>
    </div>
  );
}
