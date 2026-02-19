-- VMs (registered agents)
CREATE TABLE vms (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name           VARCHAR(255) NOT NULL UNIQUE,
  address        VARCHAR(255) NOT NULL,          -- "http://10.0.0.1:9000"
  auth_token     VARCHAR(512) NOT NULL,          -- plaintext machine credential
  labels         JSONB DEFAULT '[]',
  status         VARCHAR(50) DEFAULT 'unknown',  -- "online","unreachable","unknown"
  last_heartbeat TIMESTAMP,
  created_at     TIMESTAMP DEFAULT NOW()
);

-- Apps (registered via agent config on startup)
CREATE TABLE apps (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  vm_id             UUID REFERENCES vms(id) ON DELETE CASCADE,
  name              VARCHAR(255) NOT NULL,
  type              VARCHAR(50) NOT NULL,         -- "systemd", "docker"
  environment       VARCHAR(50),                  -- "production", "dev"
  config            JSONB NOT NULL,               -- service name, env path, health check
  last_status       VARCHAR(50),                  -- "running","stopped","unhealthy"
  last_checked_at   TIMESTAMP,
  last_restarted_at TIMESTAMP,
  created_at        TIMESTAMP DEFAULT NOW(),
  UNIQUE(vm_id, name)
);

-- Audit logs
CREATE TABLE audit_logs (
  id         BIGSERIAL PRIMARY KEY,
  app_id     UUID REFERENCES apps(id) ON DELETE SET NULL,
  action     VARCHAR(100) NOT NULL,               -- "env_update", "restart"
  details    JSONB,                               -- diff for env_update, null for restart
  created_at TIMESTAMP DEFAULT NOW()
);
