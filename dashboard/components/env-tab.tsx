"use client";

import { useState, useEffect } from "react";
import { EnvMap } from "@/lib/types";

interface Props {
  appId: string;
}

type Toast = { type: "success" | "error"; message: string };

export function EnvTab({ appId }: Props) {
  const [env, setEnv] = useState<EnvMap>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [editMode, setEditMode] = useState(false);
  const [edits, setEdits] = useState<Record<string, string>>({});
  const [showDiff, setShowDiff] = useState(false);
  const [saving, setSaving] = useState(false);
  const [revealed, setRevealed] = useState<Set<string>>(new Set());
  const [toast, setToast] = useState<Toast | null>(null);

  useEffect(() => {
    fetch(`/api/apps/${appId}/env`)
      .then((r) => r.json())
      .then((data) => { setEnv(data); setLoading(false); })
      .catch((e) => { setError(String(e)); setLoading(false); });
  }, [appId]);

  function showToast(type: Toast["type"], message: string) {
    setToast({ type, message });
    setTimeout(() => setToast(null), 4000);
  }

  function enterEditMode() {
    // Pre-fill edits with current non-masked values only.
    const initial: Record<string, string> = {};
    for (const [k, v] of Object.entries(env)) {
      if (!v.masked) initial[k] = v.value;
    }
    setEdits(initial);
    setEditMode(true);
    setShowDiff(false);
  }

  function cancelEdit() {
    setEditMode(false);
    setEdits({});
    setShowDiff(false);
  }

  // Changed non-masked keys compared to original
  const changedKeys = Object.keys(edits).filter(
    (k) => edits[k] !== env[k]?.value
  );

  async function save(restart: boolean) {
    if (changedKeys.length === 0) {
      cancelEdit();
      return;
    }
    setSaving(true);
    try {
      const body = JSON.stringify(edits);
      const res = await fetch(
        `/api/apps/${appId}/env?restart=${restart}`,
        { method: "PUT", headers: { "Content-Type": "application/json" }, body }
      );
      const data = await res.json();
      if (!res.ok) throw new Error(data.error ?? `status ${res.status}`);

      // Refresh env from server
      const fresh = await fetch(`/api/apps/${appId}/env`).then((r) => r.json());
      setEnv(fresh);
      cancelEdit();
      showToast(
        "success",
        restart
          ? "Environment saved. App is restarting."
          : "Environment variables saved."
      );
    } catch (e) {
      showToast("error", "Failed to save environment variables. Original file restored.");
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return <div className="h-48 rounded-lg border border-border bg-surface animate-pulse" />;
  }

  if (error) {
    return (
      <div className="flex items-center gap-3 bg-surface border border-danger/30 rounded-lg px-4 py-3 text-sm text-text-primary">
        <span className="h-2 w-2 rounded-full bg-danger shrink-0" />
        Cannot load environment variables. Check agent connection.
      </div>
    );
  }

  const keys = Object.keys(env).sort();

  return (
    <div className="flex flex-col gap-4">
      {/* Toast */}
      {toast && (
        <div
          className={`flex items-center gap-3 rounded-lg px-4 py-3 text-sm text-text-primary border ${
            toast.type === "success"
              ? "bg-surface border-status-running/30"
              : "bg-surface border-danger/30"
          }`}
        >
          <span
            className={`h-2 w-2 rounded-full shrink-0 ${
              toast.type === "success" ? "bg-status-running" : "bg-danger"
            }`}
          />
          {toast.message}
        </div>
      )}

      {/* Diff modal */}
      {showDiff && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/80">
          <div className="bg-surface border border-border rounded-lg w-full max-w-lg mx-4">
            <div className="flex items-center justify-between px-4 py-3 border-b border-border">
              <span className="text-sm font-semibold text-text-primary">Preview Changes</span>
              <button
                onClick={() => setShowDiff(false)}
                className="text-text-muted hover:text-text-primary transition-colors text-sm"
              >
                Close
              </button>
            </div>
            <div className="p-4 space-y-1 max-h-96 overflow-y-auto">
              {changedKeys.length === 0 ? (
                <p className="text-sm text-text-muted">No changes.</p>
              ) : (
                changedKeys.map((k) => (
                  <div key={k}>
                    <div className="font-mono text-xs bg-danger-subtle border-l-2 border-status-stopped px-3 py-1 text-text-secondary">
                      - {k}={env[k]?.value ?? ""}
                    </div>
                    <div className="font-mono text-xs bg-status-running-bg border-l-2 border-status-running px-3 py-1 text-text-secondary">
                      + {k}={edits[k]}
                    </div>
                  </div>
                ))
              )}
            </div>
            <div className="flex items-center justify-end gap-3 px-4 py-3 border-t border-border">
              <button
                onClick={() => setShowDiff(false)}
                className="bg-transparent hover:bg-surface-raised text-text-secondary hover:text-text-primary border border-border text-sm font-medium px-4 py-2 rounded-md transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={() => { setShowDiff(false); save(false); }}
                disabled={saving || changedKeys.length === 0}
                className="bg-transparent hover:bg-surface-raised text-text-secondary hover:text-text-primary border border-border text-sm font-medium px-4 py-2 rounded-md transition-colors disabled:opacity-50"
              >
                Save
              </button>
              <button
                onClick={() => { setShowDiff(false); save(true); }}
                disabled={saving || changedKeys.length === 0}
                className="bg-accent hover:bg-accent-hover text-background text-sm font-medium px-4 py-2 rounded-md transition-colors disabled:opacity-50"
              >
                Save & Restart
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Header row */}
      <div className="flex items-center justify-between">
        <span className="text-sm text-text-muted">{keys.length} variables</span>
        {!editMode ? (
          <button
            onClick={enterEditMode}
            className="bg-transparent hover:bg-surface-raised text-text-secondary hover:text-text-primary border border-border text-sm font-medium px-4 py-2 rounded-md transition-colors"
          >
            Edit
          </button>
        ) : (
          <div className="flex items-center gap-2">
            <button
              onClick={cancelEdit}
              className="bg-transparent hover:bg-surface-raised text-text-secondary hover:text-text-primary border border-border text-sm font-medium px-4 py-2 rounded-md transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={() => setShowDiff(true)}
              disabled={changedKeys.length === 0}
              className="bg-transparent hover:bg-surface-raised text-text-secondary hover:text-text-primary border border-border text-sm font-medium px-4 py-2 rounded-md transition-colors disabled:opacity-40"
            >
              Preview Changes
            </button>
            <button
              onClick={() => save(false)}
              disabled={saving || changedKeys.length === 0}
              className="bg-transparent hover:bg-surface-raised text-text-secondary hover:text-text-primary border border-border text-sm font-medium px-4 py-2 rounded-md transition-colors disabled:opacity-40"
            >
              {saving ? "Saving..." : "Save"}
            </button>
            <button
              onClick={() => save(true)}
              disabled={saving || changedKeys.length === 0}
              className="bg-accent hover:bg-accent-hover text-background text-sm font-medium px-4 py-2 rounded-md transition-colors disabled:opacity-50"
            >
              {saving ? "Saving..." : "Save & Restart"}
            </button>
          </div>
        )}
      </div>

      {/* Env table */}
      {keys.length === 0 ? (
        <div className="rounded-lg border border-border p-12 text-center">
          <p className="text-text-muted">No environment variables configured.</p>
        </div>
      ) : (
        <div className="bg-surface border border-border rounded-lg overflow-hidden">
          {keys.map((key) => {
            const entry = env[key];
            const isRevealed = revealed.has(key);
            return (
              <div
                key={key}
                className="flex items-center border-b border-border last:border-0 hover:bg-surface-raised transition-colors px-4 py-3"
              >
                <span className="font-mono text-sm text-text-primary w-56 shrink-0">
                  {key}
                </span>
                {editMode && !entry.masked ? (
                  <input
                    value={edits[key] ?? ""}
                    onChange={(e) => setEdits((p) => ({ ...p, [key]: e.target.value }))}
                    className="flex-1 bg-background border border-border hover:border-border-subtle focus:border-accent rounded-md px-3 py-1 text-sm text-text-primary outline-none transition-colors font-mono"
                  />
                ) : (
                  <span className="font-mono text-sm text-text-secondary flex-1">
                    {entry.masked
                      ? isRevealed
                        ? entry.value
                        : "- - - - - - - -"
                      : entry.value}
                  </span>
                )}
                {entry.masked && !editMode && (
                  <button
                    onClick={() =>
                      setRevealed((prev) => {
                        const next = new Set(prev);
                        next.has(key) ? next.delete(key) : next.add(key);
                        return next;
                      })
                    }
                    className="text-xs text-text-muted hover:text-accent transition-colors ml-3"
                  >
                    {isRevealed ? "hide" : "reveal"}
                  </button>
                )}
                {entry.masked && editMode && (
                  <span className="text-xs text-text-disabled ml-3">masked</span>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
