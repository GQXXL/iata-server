#!/usr/bin/env bash
set -euo pipefail

if [ "$(id -u)" -ne 0 ]; then
  echo "[fatal] Please run as root"
  exit 1
fi

SERVICE="ppanel-network-activity.service"
ENV_FILE="/etc/ppanel/network-activity.env"
SCRIPT_FILE="/opt/report_network_activity_journal.sh"
UNIT_FILE="/etc/systemd/system/${SERVICE}"

echo "[1/4] stop & disable service"
systemctl stop "$SERVICE" 2>/dev/null || true
systemctl disable "$SERVICE" 2>/dev/null || true

echo "[2/4] remove files"
for f in "$UNIT_FILE" "$ENV_FILE" "$SCRIPT_FILE"; do
  if [ -f "$f" ]; then
    rm -f "$f"
    echo "removed: $f"
  fi
done

echo "[3/4] reload systemd"
systemctl daemon-reload || true
systemctl reset-failed "$SERVICE" 2>/dev/null || true

echo "[4/4] verify"
if systemctl list-unit-files | grep -q "^${SERVICE}"; then
  echo "[warn] unit file entry still exists (may be cached), run: systemctl daemon-reexec"
fi

echo "[done] network activity reporter uninstalled"
