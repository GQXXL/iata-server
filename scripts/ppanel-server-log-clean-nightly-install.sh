#!/usr/bin/env bash
set -euo pipefail

# Required params (can be passed inline before command)
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-ppanel}"
DB_PASS="${DB_PASS:-}"
DB_NAME="${DB_NAME:-ppanel}"

if [ "$(id -u)" -ne 0 ]; then
  echo "[fatal] Please run as root"
  exit 1
fi

if [ -z "$DB_PASS" ]; then
  echo "[fatal] DB_PASS is empty"
  echo "usage: DB_PASS='xxx' [DB_HOST=127.0.0.1] [DB_PORT=3306] [DB_USER=ppanel] [DB_NAME=ppanel] bash <(curl -sL <script_url>)"
  exit 1
fi

echo "[1/5] write cleanup script"
cat >/usr/local/bin/ppanel-log-clean-nightly.sh <<'SH'
#!/usr/bin/env bash
set -euo pipefail

DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-ppanel}"
DB_PASS="${DB_PASS:-}"
DB_NAME="${DB_NAME:-ppanel}"

if [ -z "$DB_PASS" ]; then
  echo "[fatal] DB_PASS is empty"
  exit 1
fi

# 1) clean network activity before today
mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" <<'SQL'
DELETE FROM user_network_activity
WHERE `timestamp` < CURDATE();
SQL

# 2) keep subscribe logs (system_log type=20) -> do nothing

# 3) journald cleanup
systemctl kill -s SIGUSR2 systemd-journald 2>/dev/null || true
journalctl --vacuum-time=1s 2>/dev/null || true

# 4) execution log
mkdir -p /var/log
printf '%s nightly-clean done\n' "$(date '+%F %T')" >> /var/log/ppanel-nightly-clean.log
SH
chmod +x /usr/local/bin/ppanel-log-clean-nightly.sh

echo "[2/5] write env"
mkdir -p /etc/ppanel
cat >/etc/ppanel/server-log-clean-nightly.env <<ENV
DB_HOST=${DB_HOST}
DB_PORT=${DB_PORT}
DB_USER=${DB_USER}
DB_PASS=${DB_PASS}
DB_NAME=${DB_NAME}
ENV
chmod 600 /etc/ppanel/server-log-clean-nightly.env

echo "[3/5] wire env into cron"
cat >/etc/cron.d/ppanel-log-clean-nightly <<'CRON'
SHELL=/bin/bash
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

# every day at 00:00
0 0 * * * root . /etc/ppanel/server-log-clean-nightly.env; /usr/local/bin/ppanel-log-clean-nightly.sh
CRON
chmod 644 /etc/cron.d/ppanel-log-clean-nightly

echo "[4/5] restart cron"
systemctl restart cron 2>/dev/null || systemctl restart crond 2>/dev/null || true

echo "[5/5] status"
systemctl status cron --no-pager -l | sed -n '1,20p' || systemctl status crond --no-pager -l | sed -n '1,20p' || true

echo "[ok] nightly cleanup installed (network activity + journald only)"
