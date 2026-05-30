# Network Activity Reporter（一键安装/卸载）

## 1) 一键安装（节点机执行）

> 请先替换 `SERVER_ID`、`SECRET_KEY`、`API_BASE`

```bash
SERVER_ID="1" SECRET_KEY="你的密钥" API_BASE="https://你的域名或IP:端口" PROTOCOL="xray" bash <(curl -sL https://raw.githubusercontent.com/GQXXL/iata-server/Network-Activity/scripts/ppanel-network-activity-install.sh)
```

- `SERVER_ID`：对应节点的服务器 ID（数字）
- `SECRET_KEY`：该节点鉴权密钥
- `API_BASE`：后端地址，支持域名/IP（如 `https://panel.example.com` 或 `http://1.2.3.4:8081`）
- `PROTOCOL`：默认 `xray`

## 2) 状态检查

```bash
systemctl status ppanel-network-activity.service --no-pager -l
journalctl -u ppanel-network-activity.service -f --no-pager
```

## 3) 一键卸载（节点机执行）

```bash
bash <(curl -sL https://raw.githubusercontent.com/GQXXL/iata-server/Network-Activity/scripts/ppanel-network-activity-uninstall.sh)
```

## 4) 卸载后检查

```bash
systemctl status ppanel-network-activity.service --no-pager -l || true
ls -l /etc/ppanel/network-activity.env /opt/report_network_activity_journal.sh /etc/systemd/system/ppanel-network-activity.service 2>/dev/null || true
```
