#!/usr/bin/env bash
set -euo pipefail

if [ "$(id -u)" -ne 0 ]; then
  echo "[fatal] Please run as root"
  exit 1
fi

CRON_FILE="/etc/cron.d/ppanel-log-clean-nightly"
SCRIPT_FILE="/usr/local/bin/ppanel-log-clean-nightly.sh"
ENV_FILE="/etc/ppanel/server-log-clean-nightly.env"

echo "[1/3] remove files"
[ -f "$CRON_FILE" ] && rm -f "$CRON_FILE" && echo "removed: $CRON_FILE" || true
[ -f "$SCRIPT_FILE" ] && rm -f "$SCRIPT_FILE" && echo "removed: $SCRIPT_FILE" || true
[ -f "$ENV_FILE" ] && rm -f "$ENV_FILE" && echo "removed: $ENV_FILE" || true

echo "[2/3] restart cron"
systemctl restart cron 2>/dev/null || systemctl restart crond 2>/dev/null || true

echo "[3/3] done"
echo "[ok] server nightly cleanup uninstalled"
