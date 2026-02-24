export type VMStatus = "online" | "unreachable" | "unknown";
export type AppStatus = "running" | "stopped" | "unhealthy" | "";

export interface AppConfig {
  service?: string;
  container?: string;
  env_file?: string;
  auto_restart?: boolean;
  deploy_dir?: string;
}

export interface App {
  id: string;
  vm_id: string;
  vm_name: string;
  name: string;
  type: string;
  environment?: string;
  config: AppConfig;
  last_status: AppStatus;
  last_checked_at?: string;
  last_restarted_at?: string;
  created_at: string;
}

export interface DeployResult {
  success: boolean;
  output: string;
  error?: string;
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
  vm_peak_mb: number;
  pid: number;
  sampled_at: string;
}

export interface SystemMetrics {
  mem_total_mb: number;
  mem_used_mb: number;
  mem_free_mb: number;
  load_avg_1: number;
  load_avg_5: number;
  load_avg_15: number;
  uptime_seconds: number;
  disk_total_gb: number;
  disk_used_gb: number;
  disk_free_gb: number;
  sampled_at: string;
}

export interface AuditLog {
  id: number;
  app_id: string | null;
  action: string;
  details: Record<string, unknown> | null;
  created_at: string;
}
