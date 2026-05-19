#!/bin/bash
# OKX Monitor 一键安装脚本（Debian/Ubuntu）
# 用法: curl -sSL https://raw.githubusercontent.com/9623739-bot/okx-monitor/master/deploy.sh | bash

set -e

APP_DIR="/opt/okx-monitor"
SERVICE_NAME="okx-monitor"
PORT=8088
REPO="https://raw.githubusercontent.com/9623739-bot/okx-monitor/master"

echo "========================================"
echo "  OKX Monitor 一键安装"
echo "========================================"

# 1. 创建目录
echo "[1/3] 创建目录..."
mkdir -p "$APP_DIR/frontend/css" "$APP_DIR/frontend/js"

# 2. 下载预编译二进制 + 前端
echo "[2/3] 下载程序..."
curl -sSL "$REPO/okx-monitor-linux" -o "$APP_DIR/okx-monitor"
chmod +x "$APP_DIR/okx-monitor"

curl -sSL "$REPO/frontend/css/base.css"    -o "$APP_DIR/frontend/css/base.css"
curl -sSL "$REPO/frontend/js/common.js"    -o "$APP_DIR/frontend/js/common.js"
curl -sSL "$REPO/frontend/index.html"      -o "$APP_DIR/frontend/index.html"
curl -sSL "$REPO/frontend/dashboard.html"  -o "$APP_DIR/frontend/dashboard.html"
curl -sSL "$REPO/frontend/monitor.html"    -o "$APP_DIR/frontend/monitor.html"
curl -sSL "$REPO/frontend/history.html"    -o "$APP_DIR/frontend/history.html"
curl -sSL "$REPO/frontend/settings.html"   -o "$APP_DIR/frontend/settings.html"

# 3. 创建 systemd 服务
echo "[3/3] 创建系统服务..."
cat > /etc/systemd/system/$SERVICE_NAME.service << EOF
[Unit]
Description=OKX Monitor - 合约异动监控交易系统
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/okx-monitor
Restart=always
RestartSec=5
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable $SERVICE_NAME
systemctl start $SERVICE_NAME

echo ""
echo "========================================"
echo "  安装完成！"
echo "========================================"
echo "  访问地址: http://$(curl -s ifconfig.me):$PORT"
echo "  登录账号: admin / admin123"
echo ""
echo "  管理命令:"
echo "    systemctl status $SERVICE_NAME  # 状态"
echo "    systemctl restart $SERVICE_NAME # 重启"
echo "    journalctl -u $SERVICE_NAME -f  # 日志"
echo "========================================"
