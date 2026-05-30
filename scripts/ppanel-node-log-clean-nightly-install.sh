#!/usr/bin/env bash
set -euo pipefail

if [ "$(id -u)" -ne 0 ]; then
  echo "[fatal] Please run as root"
  exit 1
fi

echo "[1/4] write cleanup script"
cat >/usr/local/bin/ppanel-node-log-clean-nightly.sh <<'SH'
#!/usr/bin/env bash
set -euo pipefail

# trigger journald rotation (compatible fallback)
systemctl kill -s SIGUSR2 systemd-journald 2>/dev/null || true

# aggressively vacuum old logs
journalctl --vacuum-time=1s 2>/dev/null || true

# record execution
mkdir -p /var/log
printf '%s node-nightly-clean done\n' "$(date '+%F %T')" >> /var/log/ppanel-node-nightly-clean.log
SH
chmod +x /usr/local/bin/ppanel-node-log-clean-nightly.sh

echo "[2/4] write cron job"
cat >/etc/cron.d/ppanel-node-log-clean-nightly <<'CRON'
SHELL=/bin/bash
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

# run daily at 00:00
0 0 * * * root /usr/local/bin/ppanel-node-log-clean-nightly.sh
CRON
chmod 644 /etc/cron.d/ppanel-node-log-clean-nightly

echo "[3/4] restart cron"
systemctl restart cron 2>/dev/null || systemctl restart crond 2>/dev/null || true

echo "[4/4] status"
systemctl status cron --no-pager -l | sed -n '1,20p' || systemctl status crond --no-pager -l | sed -n '1,20p' || true

echo "[ok] node log cleanup scheduled at 00:00 daily"
