# Network Activity Reporter（统一安装/卸载手册）

本文统一说明 3 组脚本：
1) 节点机网络活动上报（install/uninstall）
2) 节点机日志清理（install/uninstall）
3) 服务端数据库+journald 清理（install/uninstall，仅清网络活动，不清订阅日志）

---

## 一、节点机网络活动上报

### 1) 一键安装（节点机执行）

> 请先替换 `SERVER_ID`、`SECRET_KEY`、`API_BASE`

```bash
SERVER_ID="1" SECRET_KEY="你的密钥" API_BASE="https://你的域名或IP:端口" PROTOCOL="xray" bash <(curl -sL https://raw.githubusercontent.com/GQXXL/iata-server/Network-Activity/scripts/ppanel-network-activity-install.sh)
```

- `SERVER_ID`：对应节点的服务器 ID（数字）
- `SECRET_KEY`：该节点鉴权密钥
- `API_BASE`：后端地址，支持域名/IP（如 `https://panel.example.com` 或 `http://1.2.3.4:8081`）
- `PROTOCOL`：默认 `xray`

### 2) 状态检查

```bash
systemctl status ppanel-network-activity.service --no-pager -l
journalctl -u ppanel-network-activity.service -f --no-pager
```

### 3) 一键卸载（节点机执行）

```bash
bash <(curl -sL https://raw.githubusercontent.com/GQXXL/iata-server/Network-Activity/scripts/ppanel-network-activity-uninstall.sh)
```

### 4) 卸载后检查

```bash
systemctl status ppanel-network-activity.service --no-pager -l || true
ls -l /etc/ppanel/network-activity.env /opt/report_network_activity_journal.sh /etc/systemd/system/ppanel-network-activity.service 2>/dev/null || true
```

---

## 二、节点机日志清理（每天 00:00）

### 1) 一键安装（节点机执行）

```bash
bash <(curl -sL https://raw.githubusercontent.com/GQXXL/iata-server/Network-Activity/scripts/ppanel-node-log-clean-nightly-install.sh)
```

功能：
- 每天 `00:00` 执行一次
- 轮转 journald
- `journalctl --vacuum-time=1s` 清理历史日志
- 写入执行记录：`/var/log/ppanel-node-nightly-clean.log`

### 2) 状态检查

```bash
cat /etc/cron.d/ppanel-node-log-clean-nightly
systemctl status cron --no-pager -l || systemctl status crond --no-pager -l
```

### 3) 一键卸载（节点机执行）

```bash
bash <(curl -sL https://raw.githubusercontent.com/GQXXL/iata-server/Network-Activity/scripts/ppanel-node-log-clean-nightly-uninstall.sh)
```

### 4) 卸载后检查

```bash
ls -l /etc/cron.d/ppanel-node-log-clean-nightly /usr/local/bin/ppanel-node-log-clean-nightly.sh 2>/dev/null || true
```

---

## 三、服务端清理（每天 00:00，仅网络活动 + journald）

> 说明：只清理 `user_network_activity` 历史数据和系统 journal，**不清理订阅日志**（`system_log type=20`）。

### 1) 一键安装（服务端执行）

```bash
DB_PASS='你的数据库密码' DB_HOST='127.0.0.1' DB_PORT='3306' DB_USER='ppanel' DB_NAME='ppanel' bash <(curl -sL https://raw.githubusercontent.com/GQXXL/iata-server/Network-Activity/scripts/ppanel-server-log-clean-nightly-install.sh)
```

### 2) 状态检查

```bash
cat /etc/cron.d/ppanel-log-clean-nightly
cat /etc/ppanel/server-log-clean-nightly.env
tail -n 20 /var/log/ppanel-nightly-clean.log 2>/dev/null || true
systemctl status cron --no-pager -l || systemctl status crond --no-pager -l
```

### 3) 一键卸载（服务端执行）

```bash
bash <(curl -sL https://raw.githubusercontent.com/GQXXL/iata-server/Network-Activity/scripts/ppanel-server-log-clean-nightly-uninstall.sh)
```

### 4) 卸载后检查

```bash
ls -l /etc/cron.d/ppanel-log-clean-nightly /usr/local/bin/ppanel-log-clean-nightly.sh /etc/ppanel/server-log-clean-nightly.env 2>/dev/null || true
```

---

## 常见注意事项

- 全部脚本都建议 root 执行。
- `API_BASE` 支持域名或 IP，只要节点机能访问即可。
- Debian/Ubuntu 通常是 `cron`，CentOS/RHEL 可能是 `crond`，脚本已兼容两者重启方式。
