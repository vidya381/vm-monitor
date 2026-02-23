export type VMStatus = "online" | "unreachable" | "unknown";
export type AppStatus = "running" | "stopped" | "unhealthy" | "";

export interface App {
  id: string;
  vm_id: string;
  vm_name: string;
  name: string;
  type: string;
  environment?: string;
  last_status: AppStatus;
  last_checked_at?: string;
  last_restarted_at?: string;
  created_at: string;
}

export interface VM {
  id: string;
  name: string;
  labels: string[];
  status: VMStatus;
  last_heartbeat?: string;
  created_at: string;
}

export interface LogResult {
  lines: string[];
  cursor: string;
  has_more: boolean;
}

export interface EnvEntry {
  value: string;
  masked: boolean;
}

export type EnvMap = Record<string, EnvEntry>;

export interface Metrics {
  cpu_percent: number;
  mem_rss_mb: number;
  pid: number;
  sampled_at: string;
}

export interface AuditLog {
  id: number;
  app_id: string | null;
  action: string;
  details: Record<string, unknown> | null;
  created_at: string;
}
