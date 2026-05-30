#!/usr/bin/env bash
set -euo pipefail

if [ "$(id -u)" -ne 0 ]; then
  echo "[fatal] Please run as root"
  exit 1
fi

CRON_FILE="/etc/cron.d/ppanel-node-log-clean-nightly"
SCRIPT_FILE="/usr/local/bin/ppanel-node-log-clean-nightly.sh"

echo "[1/3] remove files"
[ -f "$CRON_FILE" ] && rm -f "$CRON_FILE" && echo "removed: $CRON_FILE" || true
[ -f "$SCRIPT_FILE" ] && rm -f "$SCRIPT_FILE" && echo "removed: $SCRIPT_FILE" || true

echo "[2/3] reload cron"
systemctl restart cron 2>/dev/null || systemctl restart crond 2>/dev/null || true

echo "[3/3] done"
echo "[ok] nightly log cleanup uninstalled"
