// All fake data used in demo mode. Never touches the real control plane.

export const isDemoMode = () => process.env.DEMO_MODE === "true";

const now = () => new Date().toISOString();
const minsAgo = (m: number) => new Date(Date.now() - m * 60_000).toISOString();
const daysAgo = (d: number) => new Date(Date.now() - d * 86_400_000).toISOString();

export const DEMO_BLOCKED = { error: "Demo mode — changes are disabled" };

// --- VMs ---

export const demoVMs = [
  {
    id: "demo-vm-1",
    name: "oracle-amd1",
    status: "online",
    labels: ["production", "oracle-free-tier"],
    last_heartbeat: minsAgo(1),
    created_at: daysAgo(30),
  },
  {
    id: "demo-vm-2",
    name: "oracle-amd2",
    status: "online",
    labels: ["production", "oracle-free-tier"],
    last_heartbeat: minsAgo(0),
    created_at: daysAgo(30),
  },
];

// --- Apps ---

export const demoApps = [
  {
    id: "demo-app-1",
    vm_id: "demo-vm-1",
    vm_name: "oracle-amd1",
    name: "myspendo",
    type: "systemd",
    environment: "production",
    config: { service: "myspendo.service", deploy_dir: "/home/ubuntu/myspendo-backend", auto_restart: true },
    last_status: "running",
    last_checked_at: minsAgo(1),
    last_restarted_at: daysAgo(3),
    created_at: daysAgo(30),
  },
  {
    id: "demo-app-2",
    vm_id: "demo-vm-1",
    vm_name: "oracle-amd1",
    name: "weather-insight",
    type: "systemd",
    environment: "production",
    config: { service: "weather-insight.service", deploy_dir: "/home/ubuntu/weather-insight-backend", auto_restart: true },
    last_status: "running",
    last_checked_at: minsAgo(1),
    last_restarted_at: daysAgo(7),
    created_at: daysAgo(30),
  },
  {
    id: "demo-app-4",
    vm_id: "demo-vm-2",
    vm_name: "oracle-amd2",
    name: "vm-monitor-api",
    type: "systemd",
    environment: "production",
    config: { service: "vm-monitor-api.service", deploy_dir: "/home/ubuntu/vm-monitor-api" },
    last_status: "running",
    last_checked_at: minsAgo(0),
    last_restarted_at: daysAgo(1),
    created_at: daysAgo(30),
  },
];

// --- Logs ---

export const demoLogs: Record<string, string[]> = {
  "demo-app-1": [
    `${now()} INFO  server started on :8080`,
    `${now()} INFO  connected to database`,
    `${minsAgo(10)} INFO  GET /api/health 200 2ms`,
    `${minsAgo(8)} INFO  POST /api/transactions 201 18ms`,
    `${minsAgo(5)} INFO  GET /api/transactions 200 14ms`,
    `${minsAgo(3)} WARN  slow query detected: 243ms`,
    `${minsAgo(1)} INFO  GET /api/health 200 1ms`,
  ],
  "demo-app-2": [
    `${now()} INFO  weather service started`,
    `${minsAgo(15)} INFO  fetched forecast for Bengaluru, IN`,
    `${minsAgo(10)} INFO  scheduled refresh completed in 142ms`,
    `${minsAgo(5)} INFO  GET /api/weather 200 8ms`,
    `${minsAgo(2)} INFO  GET /api/forecast 200 12ms`,
  ],
  "demo-app-4": [
    `${now()} INFO  starting vm-monitor api addr=:9090`,
    `${minsAgo(2)} INFO  registered agent oracle-amd1 apps=2`,
    `${minsAgo(1)} INFO  poll cycle completed vms=2 apps=3`,
    `${minsAgo(0)} INFO  GET /health 200`,
  ],
};

// --- Metrics ---

export const demoMetrics: Record<string, object> = {
  "demo-app-1": { cpu_percent: 0.4, mem_rss_mb: 38.2, vm_peak_mb: 52.1, pid: 1234, sampled_at: now() },
  "demo-app-2": { cpu_percent: 0.1, mem_rss_mb: 124.6, vm_peak_mb: 148.3, pid: 1891, sampled_at: now() },
  "demo-app-4": { cpu_percent: 0.2, mem_rss_mb: 41.1, vm_peak_mb: 55.8, pid: 892, sampled_at: now() },
};

export const demoSystemMetrics: Record<string, object> = {
  "demo-vm-1": {
    mem_total_mb: 23980.0, mem_used_mb: 3241.4, mem_free_mb: 20738.6,
    load_avg_1: 0.12, load_avg_5: 0.09, load_avg_15: 0.07,
    uptime_seconds: 1_382_400,
    disk_total_gb: 46.6, disk_used_gb: 12.3, disk_free_gb: 34.3,
    sampled_at: now(),
  },
  "demo-vm-2": {
    mem_total_mb: 23980.0, mem_used_mb: 1876.2, mem_free_mb: 22103.8,
    load_avg_1: 0.05, load_avg_5: 0.06, load_avg_15: 0.04,
    uptime_seconds: 2_764_800,
    disk_total_gb: 46.6, disk_used_gb: 8.7, disk_free_gb: 37.9,
    sampled_at: now(),
  },
};

// --- Env vars ---

export const demoEnvFiles: Record<string, string[]> = {
  "demo-app-1": [".env"],
};

export const demoEnvVars: Record<string, object> = {
  "demo-app-1": {
    PORT:         { value: "8080",       masked: false },
    APP_ENV:      { value: "production", masked: false },
    LOG_LEVEL:    { value: "info",       masked: false },
    DATABASE_URL: { value: "••••••••",   masked: true  },
    JWT_SECRET:   { value: "••••••••",   masked: true  },
  },
};

// --- Audit ---

// --- Uptime ---

export const demoUptime: Record<string, object> = {
  "demo-app-1": {
    app_id: "demo-app-1",
    window_days: 30,
    uptime_pct: 99.2,
    total_downtime_s: 13320,
    incident_count: 2,
    incidents: [
      { status: "stopped",   started_at: daysAgo(2), ended_at: new Date(Date.now() - 2 * 86_400_000 + 7800 * 1000).toISOString(), duration_s: 7800 },
      { status: "unhealthy", started_at: daysAgo(6), ended_at: new Date(Date.now() - 6 * 86_400_000 + 5520 * 1000).toISOString(), duration_s: 5520 },
    ],
  },
  "demo-app-2": {
    app_id: "demo-app-2",
    window_days: 30,
    uptime_pct: 100.0,
    total_downtime_s: 0,
    incident_count: 0,
    incidents: [],
  },
  "demo-app-4": {
    app_id: "demo-app-4",
    window_days: 30,
    uptime_pct: 99.8,
    total_downtime_s: 5184,
    incident_count: 1,
    incidents: [
      { status: "stopped", started_at: daysAgo(1), ended_at: new Date(Date.now() - 1 * 86_400_000 + 5184 * 1000).toISOString(), duration_s: 5184 },
    ],
  },
};

// --- Audit ---

export const demoAudit: Record<string, object[]> = {
  "demo-app-1": [
    { id: 4, app_id: "demo-app-1", action: "deploy",     created_at: daysAgo(1),  details: { output: "Already up to date." } },
    { id: 3, app_id: "demo-app-1", action: "restart",    created_at: daysAgo(3),  details: null },
    { id: 2, app_id: "demo-app-1", action: "env_update", created_at: daysAgo(5),  details: { keys_updated: 1, restart: true } },
    { id: 1, app_id: "demo-app-1", action: "app_registered", created_at: daysAgo(30), details: { name: "myspendo", type: "systemd", vm: "oracle-amd1" } },
  ],
  "demo-app-2": [
    { id: 6, app_id: "demo-app-2", action: "deploy",  created_at: daysAgo(7),  details: { output: "Updating abc123..def456\nFast-forward\n 2 files changed" } },
    { id: 5, app_id: "demo-app-2", action: "restart", created_at: daysAgo(7),  details: null },
  ],
  "demo-app-4": [
    { id: 7, app_id: "demo-app-4", action: "deploy",  created_at: daysAgo(1),  details: { output: "Already up to date." } },
    { id: 8, app_id: "demo-app-4", action: "restart", created_at: daysAgo(1),  details: null },
  ],
};
