#!/usr/bin/env bash
set -euo pipefail

SERVER_ID="${SERVER_ID:-}"
SECRET_KEY="${SECRET_KEY:-}"
API_BASE="${API_BASE:-http://38.146.25.75:8081}"
PROTOCOL="${PROTOCOL:-xray}"

if [ "$(id -u)" -ne 0 ]; then
  echo "[fatal] Please run as root"
  exit 1
fi

if ! [[ "$SERVER_ID" =~ ^[0-9]+$ ]]; then
  echo "[fatal] SERVER_ID must be numeric. current: ${SERVER_ID:-<empty>}"
  echo "usage: SERVER_ID=1 SECRET_KEY=xxxxx [API_BASE=http://ip:8081] [PROTOCOL=xray] bash <(curl -sL <this_script_url>)"
  exit 1
fi

if [ -z "$SECRET_KEY" ]; then
  echo "[fatal] SECRET_KEY is empty"
  echo "usage: SERVER_ID=1 SECRET_KEY=xxxxx [API_BASE=http://ip:8081] [PROTOCOL=xray] bash <(curl -sL <this_script_url>)"
  exit 1
fi

echo "[1/5] write reporter script"
cat >/opt/report_network_activity_journal.sh <<'SH'
#!/usr/bin/env bash
set -u

API_BASE="${PPANEL_API_BASE:-http://38.146.25.75:8081}"
SERVER_ID="${PPANEL_SERVER_ID:-1}"
PROTOCOL="${PPANEL_PROTOCOL:-xray}"
SECRET_KEY="${PPANEL_SECRET_KEY:-}"

if [ -z "${SECRET_KEY}" ]; then
  echo "[fatal] PPANEL_SECRET_KEY empty"
  exit 1
fi

URL="${API_BASE%/}/v1/server/push_network_activity?server_id=${SERVER_ID}&protocol=${PROTOCOL}&secret_key=${SECRET_KEY}"
echo "[start] API_BASE=${API_BASE} SERVER_ID=${SERVER_ID} PROTOCOL=${PROTOCOL}"

journalctl -u PPanel-node -f -n 0 --no-pager -o cat | while IFS= read -r line; do
  cip="$(echo "$line" | sed -nE 's/.*from ([0-9a-fA-F\.:]+):[0-9]+ accepted.*/\1/p')"
  dom="$(echo "$line" | sed -nE 's/.*accepted (tcp|udp):([^: ]+):[0-9]+.*/\2/p')"
  uuid="$(echo "$line" | grep -oE '[0-9a-fA-F-]{8}-[0-9a-fA-F-]{4}-[0-9a-fA-F-]{4}-[0-9a-fA-F-]{4}-[0-9a-fA-F-]{12}$' || true)"

  [ -z "${uuid:-}" ] && continue
  [ -z "${dom:-}" ] && continue

  ts="$(date +%s)"
  payload="{\"records\":[{\"user_subscribe_uuid\":\"$uuid\",\"client_ip\":\"${cip:-}\",\"domain\":\"${dom}\",\"upload\":0,\"download\":0,\"timestamp\":$ts,\"user_agent\":\"PPanel-node/journalctl-uuid-auto\"}]}"
  resp="$(curl -sS -m 10 -H 'Content-Type: application/json' -X POST "$URL" -d "$payload" || true)"

  echo "[push] uuid=$uuid dom=$dom resp=$resp"
done
SH
chmod +x /opt/report_network_activity_journal.sh

echo "[2/5] write env"
mkdir -p /etc/ppanel
cat >/etc/ppanel/network-activity.env <<ENV
PPANEL_API_BASE=${API_BASE}
PPANEL_SERVER_ID=${SERVER_ID}
PPANEL_PROTOCOL=${PROTOCOL}
PPANEL_SECRET_KEY=${SECRET_KEY}
ENV
chmod 600 /etc/ppanel/network-activity.env

echo "[3/5] write systemd service"
cat >/etc/systemd/system/ppanel-network-activity.service <<'UNIT'
[Unit]
Description=PPanel Network Activity Reporter (journalctl)
After=network.target

[Service]
Type=simple
EnvironmentFile=/etc/ppanel/network-activity.env
ExecStart=/usr/bin/bash /opt/report_network_activity_journal.sh
Restart=always
RestartSec=2

[Install]
WantedBy=multi-user.target
UNIT

echo "[4/5] start service"
systemctl daemon-reload
systemctl enable --now ppanel-network-activity.service

echo "[5/5] status"
systemctl status ppanel-network-activity.service --no-pager -l || true
journalctl -u ppanel-network-activity.service -n 20 --no-pager || true

echo
echo "[done] installed"
echo "follow logs: journalctl -u ppanel-network-activity.service -f --no-pager"
