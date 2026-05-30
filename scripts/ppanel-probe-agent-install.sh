#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-}"
TOKEN="${TOKEN:-}"

if [ "$(id -u)" -ne 0 ]; then
  echo "Please run this script as root"
  exit 1
fi

if [ -z "$BASE_URL" ] || [ -z "$TOKEN" ]; then
  echo "Usage: BASE_URL=\"https://example.com\" TOKEN=\"xxxx\" bash <(curl -sL <raw-url>)"
  echo "Error: BASE_URL and TOKEN are required"
  exit 1
fi

install_jq() {
  if command -v jq >/dev/null 2>&1; then
    return 0
  fi

  if command -v apt-get >/dev/null 2>&1; then
    apt-get update -y && apt-get install -y jq
  elif command -v dnf >/dev/null 2>&1; then
    dnf install -y jq
  elif command -v yum >/dev/null 2>&1; then
    yum install -y jq
  elif command -v apk >/dev/null 2>&1; then
    apk add --no-cache jq
  else
    echo "No supported package manager found, cannot auto-install jq"
    exit 1
  fi
}

install_jq

mkdir -p /etc/default
cat >/etc/default/ppanel-probe-agent <<CFG
BASE_URL="$BASE_URL"
TOKEN="$TOKEN"
CFG

cat >/usr/local/bin/ppanel-probe-agent.sh <<'AGENT'
#!/usr/bin/env bash
set -euo pipefail

source /etc/default/ppanel-probe-agent

while true; do
  CFG="$(curl -fsS "$BASE_URL/v1/probe_agent/config?token=$TOKEN" 2>/dev/null)" || {
    echo "[probe-agent] config failed" >&2
    sleep 10
    continue
  }

  INTERVAL="$(echo "$CFG" | jq -r '.data.interval_seconds // 10')"
  if ! [[ "$INTERVAL" =~ ^[0-9]+$ ]] || [ "$INTERVAL" -lt 1 ]; then
    INTERVAL=10
  fi

  HB="$(jq -nc --arg token "$TOKEN" --arg version "shell-mvp" '{token:$token,version:$version}')"
  curl -fsS -X POST "$BASE_URL/v1/probe_agent/heartbeat" \
    -H "Content-Type: application/json" \
    -d "$HB" >/dev/null || echo "[probe-agent] heartbeat failed" >&2

  sleep "$INTERVAL"
done
AGENT

chmod +x /usr/local/bin/ppanel-probe-agent.sh

cat >/etc/systemd/system/ppanel-probe-agent.service <<'UNIT'
[Unit]
Description=PPanel Probe Agent
After=network.target

[Service]
Type=simple
EnvironmentFile=/etc/default/ppanel-probe-agent
ExecStart=/usr/bin/env bash /usr/local/bin/ppanel-probe-agent.sh
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
UNIT

# 每小时自动清理 journald 历史日志（保留最近 1 小时）
cat >/usr/local/bin/ppanel-probe-log-clean.sh <<'CLEAN'
#!/usr/bin/env bash
set -euo pipefail
systemctl kill -s SIGUSR2 systemd-journald 2>/dev/null || true
journalctl --vacuum-time=1h 2>/dev/null || journalctl --vacuum-size=100M
CLEAN
chmod +x /usr/local/bin/ppanel-probe-log-clean.sh

cat >/etc/cron.d/ppanel-probe-log-clean <<'CRON'
SHELL=/bin/bash
PATH=/usr/sbin:/usr/bin:/sbin:/bin
0 * * * * root /usr/local/bin/ppanel-probe-log-clean.sh >/dev/null 2>&1
CRON
chmod 644 /etc/cron.d/ppanel-probe-log-clean

systemctl daemon-reload
systemctl enable --now ppanel-probe-agent
systemctl restart ppanel-probe-agent
systemctl status ppanel-probe-agent --no-pager -l

echo "Done. Config: /etc/default/ppanel-probe-agent"
