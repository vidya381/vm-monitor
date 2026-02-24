#!/usr/bin/env bash
set -euo pipefail

REPO="vidya381/vm-monitor"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/vm-monitor"
SERVICE_NAME="vm-monitor-agent"

RED='\033[0;31m'
GREEN='\033[0;32m'
BOLD='\033[1m'
NC='\033[0m'

echo ""
echo -e "  ${BOLD}VM Monitor — Agent Installer${NC}"
echo "  ============================="
echo ""

# 1. Detect architecture
ARCH=$(uname -m)
case $ARCH in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  *)
    echo -e "${RED}Unsupported architecture: $ARCH${NC}"
    exit 1
    ;;
esac

# 2. Get latest release version
echo "Fetching latest release..."
VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
  | grep '"tag_name"' | cut -d'"' -f4)

if [ -z "$VERSION" ]; then
  echo -e "${RED}Could not determine latest release version.${NC}"
  exit 1
fi

echo -e "  Latest version: ${GREEN}$VERSION${NC}"

# 3. Download binary
BINARY_URL="https://github.com/$REPO/releases/download/$VERSION/vm-monitor-agent-linux-$ARCH"
echo "Downloading agent binary..."
curl -fsSL "$BINARY_URL" -o /tmp/vm-monitor-agent
chmod +x /tmp/vm-monitor-agent
mv /tmp/vm-monitor-agent "$INSTALL_DIR/vm-monitor-agent"
echo -e "  Installed to ${GREEN}$INSTALL_DIR/vm-monitor-agent${NC}"

# 4. Collect config
echo ""
echo "Configure your agent:"
echo ""
read -rp "  Control plane URL (e.g. https://api.yourdomain.com): " CONTROL_PLANE_URL
read -rp "  Control plane API key: " CONTROL_PLANE_API_KEY
read -rp "  Agent auth token (run: openssl rand -hex 32): " AUTH_TOKEN
read -rp "  VM name [$(hostname)]: " VM_NAME
VM_NAME="${VM_NAME:-$(hostname)}"
read -rp "  Agent port [9000]: " AGENT_PORT
AGENT_PORT="${AGENT_PORT:-9000}"
read -rp "  Agent public address (e.g. http://1.2.3.4:9000): " AGENT_ADDRESS

# 5. Write config
mkdir -p "$CONFIG_DIR"
chmod 700 "$CONFIG_DIR"

cat > "$CONFIG_DIR/agent.yaml" <<EOF
vm:
  name: "$VM_NAME"
  port: $AGENT_PORT
  address: "$AGENT_ADDRESS"
  control_plane_url: "$CONTROL_PLANE_URL"
  control_plane_api_key: "$CONTROL_PLANE_API_KEY"
  auth_token: "$AUTH_TOKEN"
  labels:
    - "production"

apps: []
# Add your apps below. Example:
# - name: "myapp"
#   type: "systemd"
#   service: "myapp.service"
#   deploy_dir: "/home/ubuntu/myapp"
#   health_check:
#     type: "http"
#     url: "http://localhost:3000/health"
EOF

chmod 600 "$CONFIG_DIR/agent.yaml"
echo -e "  Config written to ${GREEN}$CONFIG_DIR/agent.yaml${NC}"

# 6. Create systemd service
cat > /etc/systemd/system/$SERVICE_NAME.service <<EOF
[Unit]
Description=VM Monitor Agent
After=network.target

[Service]
ExecStart=$INSTALL_DIR/vm-monitor-agent --config $CONFIG_DIR/agent.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=vm-monitor-agent

[Install]
WantedBy=multi-user.target
EOF

# 7. Enable and start
systemctl daemon-reload
systemctl enable "$SERVICE_NAME"
systemctl start "$SERVICE_NAME"

echo ""
echo -e "${GREEN}Agent installed and running.${NC}"
echo ""
echo "  Next steps:"
echo "  1. Edit $CONFIG_DIR/agent.yaml — add your apps"
echo "  2. sudo systemctl restart $SERVICE_NAME"
echo "  3. Check logs: journalctl -u $SERVICE_NAME -f"
echo ""
echo "  To uninstall:"
echo "  sudo systemctl stop $SERVICE_NAME && sudo systemctl disable $SERVICE_NAME"
echo "  sudo rm $INSTALL_DIR/vm-monitor-agent /etc/systemd/system/$SERVICE_NAME.service"
echo "  sudo rm -rf $CONFIG_DIR"
echo ""
