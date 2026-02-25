# VM Monitor

Self-hosted app monitoring dashboard. Deploy a lightweight agent on each VM, manage all your apps from one place.

![Go](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go&logoColor=white)
![Next.js](https://img.shields.io/badge/Next.js-16-000000?style=flat&logo=next.js&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Neon-336791?style=flat&logo=postgresql&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?style=flat&logo=typescript&logoColor=white)

**Live Demo:** https://demo-vm-monitor.vercel.app

## What it does

You run a small Go agent on each VM. The agent watches your systemd services and Docker containers, checks their health, reads their logs and env files, and reports everything back to a central API. The dashboard shows the status of all your VMs and apps in one place — with logs, metrics, uptime history, and the ability to restart or deploy apps without SSHing in.

## Features

**Monitoring**
- Status polling every 30 seconds (running / stopped / unhealthy)
- HTTP or command-based health checks
- CPU and memory metrics (RSS + peak virtual) per app
- 30-day uptime percentage with incident timeline

**Logs**
- Live log streaming via SSE
- Falls back to cursor-based HTTP polling if SSE drops
- Last 200 lines with auto-scroll

**Management**
- Restart apps from the dashboard (with confirmation)
- Deploy via `git pull` if a deploy directory is configured
- Register new apps without editing YAML manually

**Environment**
- Reads `.env` files from the VM
- Edit environment variables in place from the dashboard
- Audit log — every change is recorded with timestamp

**Alerts**
- Webhook notifications when apps go down or recover
- Supports Slack and generic JSON webhooks (Discord, ntfy, etc.)
- Includes auto-restart status in the alert message

**Demo Mode**
- Set `DEMO_MODE=true` on the dashboard to run with fake data
- All write operations are blocked; useful for showing the dashboard publicly

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│              Dashboard (Next.js — Vercel)                    │
│        VMs · Apps · Logs · Env · Audit · Uptime              │
└──────────────────────┬───────────────────────────────────────┘
                       │ REST API (API key auth)
┌──────────────────────▼───────────────────────────────────────┐
│              Control Plane API (Go)                          │
│  Polls agents every 30s · Stores history · Proxies requests  │
└────────────┬─────────────────────────┬───────────────────────┘
             │                         │
┌────────────▼──────────┐   ┌──────────▼───────────────────────┐
│  PostgreSQL (Neon)    │   │  Agent (Go) — one per VM         │
│  Apps · VMs · Uptime  │   │  Systemd · Docker · Logs · Env   │
└───────────────────────┘   └──────────────────────────────────┘
```

## Tech Stack

**Agent** — Go 1.24, chi, Viper (config), reads `/proc` for metrics

**API** — Go 1.24, chi, pgx (PostgreSQL driver), APScheduler-style ticker poller

**Dashboard** — Next.js 16, React 19, TypeScript, Tailwind CSS v4, Lucide icons

## Quick Install (Agent)

Run this on any Linux VM:

```bash
curl -fsSL https://raw.githubusercontent.com/vidya381/vm-monitor/main/scripts/install.sh | sudo bash
```

Detects your architecture (amd64/arm64), downloads the latest release binary, writes the config to `/etc/vm-monitor/agent.yaml`, and sets up a systemd service.

After install, edit the config to add your apps:

```yaml
# /etc/vm-monitor/agent.yaml

vm:
  name: "my-vm"
  port: 9000
  address: "http://1.2.3.4:9000"
  control_plane_url: "https://api.yourdomain.com"
  control_plane_api_key: "your-api-key"
  auth_token: "your-agent-token"
  labels:
    - "production"

apps:
  - name: "myapp"
    type: "systemd"
    service: "myapp.service"
    deploy_dir: "/home/ubuntu/myapp"
    env_file: "/home/ubuntu/myapp/.env"
    health_check:
      type: "http"
      url: "http://localhost:3000/health"
```

Then restart the agent:

```bash
sudo systemctl restart vm-monitor-agent
```

## Project Structure

```
vm-monitor/
├── agent/                    # Go agent — runs on each VM
│   ├── cmd/agent/            # Entrypoint
│   └── internal/
│       ├── config/           # YAML config + runtime write-back
│       ├── server/           # HTTP handlers (status, logs, env, metrics)
│       ├── health/           # HTTP and command health checks
│       ├── metrics/          # /proc CPU and memory reads
│       ├── systemd/          # systemctl restart/status
│       ├── docker/           # Docker container management
│       └── env/              # .env file reads and writes
├── api/                      # Go control plane
│   ├── cmd/api/              # Entrypoint
│   └── internal/
│       ├── handler/          # REST endpoints
│       ├── poller/           # 30s agent polling + uptime recording
│       ├── db/               # pgx queries (apps, VMs, status history)
│       ├── agentclient/      # HTTP client for proxying to agents
│       ├── notify/           # Webhook alerts (Slack + generic)
│       └── middleware/       # API key auth, CORS
├── dashboard/                # Next.js frontend
│   ├── app/                  # Pages + API routes (server-side proxy)
│   ├── components/           # App card, log viewer, env tab, audit tab
│   └── lib/                  # Types, demo data, server-side fetch
└── scripts/
    └── install.sh            # One-liner agent installer
```

## Local Setup

### Prerequisites

- Go 1.24+
- Node.js 18+
- PostgreSQL 14+ (or a free Neon project)

### 1. Agent

```bash
cd agent
cp agent.yaml.example agent.yaml
# Edit agent.yaml — set control_plane_url and auth_token
go run ./cmd/agent --config agent.yaml
```

Runs on `http://localhost:9000`

### 2. API

```bash
cd api
cp .env.example .env
# Edit .env — add DATABASE_URL and API_KEY
go run ./cmd/api
```

Runs on `http://localhost:8080`

The API creates the database tables on startup (no separate migration step needed).

### 3. Dashboard

```bash
cd dashboard
cp .env.example .env.local
# Edit .env.local — set API_URL, API_KEY, DASHBOARD_PASSWORD
npm install
npm run dev
```

Runs on `http://localhost:3000`

### Run Tests

```bash
cd dashboard && npm test
```

## Environment Variables

**API (`api/.env`)**

```bash
DATABASE_URL=postgresql://user:password@host/dbname?sslmode=require
API_KEY=your-secret-key-here
ALLOWED_ORIGINS=https://your-dashboard.vercel.app
PORT=8080

# Optional: webhook notifications (leave blank to disable)
NOTIFY_WEBHOOK_URL=https://hooks.slack.com/services/...
NOTIFY_WEBHOOK_TYPE=slack   # "slack" or "generic"
```

**Dashboard (`dashboard/.env.local`)**

```bash
API_URL=http://localhost:8080
API_KEY=your-secret-key-here
DASHBOARD_PASSWORD=changeme
SESSION_SECRET=random-string-here

# Demo mode — returns fake data, blocks all writes
DEMO_MODE=false
NEXT_PUBLIC_DEMO_MODE=false
```

## Deployment

- **Agent** — binary on the VM via the install script, or build from source with `go build ./cmd/agent`
- **API** — any Go hosting (Render, Railway, VPS). Set `ALLOWED_ORIGINS` to your dashboard URL.
- **Dashboard** — Vercel. Set all environment variables in the Vercel project settings.

GitHub Actions builds release binaries for linux/amd64 and linux/arm64 on every `v*.*.*` tag.

---

Built with Go, Next.js, and PostgreSQL. Agent is a single binary with no runtime dependencies.
