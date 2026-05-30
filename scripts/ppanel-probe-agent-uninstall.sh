#!/usr/bin/env bash
set -euo pipefail

if [ "$(id -u)" -ne 0 ]; then
  echo "[uninstall] Please run as root"
  exit 1
fi

SERVICE="ppanel-probe-agent.service"
TIMER_UNIT="ppanel-probe-log-clean.timer"
CRON_FILE="/etc/cron.d/ppanel-probe-log-clean"
ENV_FILE="/etc/default/ppanel-probe-agent"
AGENT_BIN="/usr/local/bin/ppanel-probe-agent.sh"
CLEAN_BIN="/usr/local/bin/ppanel-probe-log-clean.sh"
UNIT_FILE="/etc/systemd/system/${SERVICE}"

echo "[uninstall] Stopping and disabling service..."
systemctl stop "$SERVICE" 2>/dev/null || true
systemctl disable "$SERVICE" 2>/dev/null || true

# Optional timer cleanup (if ever introduced)
systemctl stop "$TIMER_UNIT" 2>/dev/null || true
systemctl disable "$TIMER_UNIT" 2>/dev/null || true

if [ -f "$UNIT_FILE" ]; then
  rm -f "$UNIT_FILE"
  echo "[uninstall] Removed: $UNIT_FILE"
fi

if [ -f "$CRON_FILE" ]; then
  rm -f "$CRON_FILE"
  echo "[uninstall] Removed: $CRON_FILE"
fi

for f in "$ENV_FILE" "$AGENT_BIN" "$CLEAN_BIN"; do
  if [ -f "$f" ]; then
    rm -f "$f"
    echo "[uninstall] Removed: $f"
  fi
done

systemctl daemon-reload || true
systemctl reset-failed "$SERVICE" 2>/dev/null || true

echo "[uninstall] Done."
