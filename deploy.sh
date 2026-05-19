#!/bin/bash
# OKX Monitor 一键安装脚本（Debian/Ubuntu）
# 用法: curl -sSL https://raw.githubusercontent.com/9623739-bot/okx-monitor/master/deploy.sh | bash

set -e

APP_DIR="/opt/okx-monitor"
SERVICE_NAME="okx-monitor"
PORT=8088

echo "========================================"
echo "  OKX Monitor 一键安装"
echo "========================================"

# 1. 创建目录
echo "[1/5] 创建应用目录..."
mkdir -p "$APP_DIR/frontend/css" "$APP_DIR/frontend/js"

# 2. 下载文件
echo "[2/5] 下载程序文件..."
REPO="https://raw.githubusercontent.com/9623739-bot/okx-monitor/master"

curl -sSL "$REPO/go.mod" -o "$APP_DIR/go.mod"
curl -sSL "$REPO/go.sum" -o "$APP_DIR/go.sum"
curl -sSL "$REPO/main.go" -o "$APP_DIR/main.go"
curl -sSL "$REPO/.gitignore" -o "$APP_DIR/.gitignore"

mkdir -p "$APP_DIR/config" "$APP_DIR/models" "$APP_DIR/handlers" "$APP_DIR/services"
curl -sSL "$REPO/config/config.go" -o "$APP_DIR/config/config.go"
curl -sSL "$REPO/models/models.go" -o "$APP_DIR/models/models.go"
curl -sSL "$REPO/handlers/auth.go" -o "$APP_DIR/handlers/auth.go"
curl -sSL "$REPO/handlers/monitor.go" -o "$APP_DIR/handlers/monitor.go"
curl -sSL "$REPO/handlers/settings.go" -o "$APP_DIR/handlers/settings.go"
curl -sSL "$REPO/handlers/sse.go" -o "$APP_DIR/handlers/sse.go"
curl -sSL "$REPO/services/monitor.go" -o "$APP_DIR/services/monitor.go"
curl -sSL "$REPO/services/okx_ws.go" -o "$APP_DIR/services/okx_ws.go"

curl -sSL "$REPO/frontend/css/base.css" -o "$APP_DIR/frontend/css/base.css"
curl -sSL "$REPO/frontend/js/common.js" -o "$APP_DIR/frontend/js/common.js"
curl -sSL "$REPO/frontend/index.html" -o "$APP_DIR/frontend/index.html"
curl -sSL "$REPO/frontend/dashboard.html" -o "$APP_DIR/frontend/dashboard.html"
curl -sSL "$REPO/frontend/monitor.html" -o "$APP_DIR/frontend/monitor.html"
curl -sSL "$REPO/frontend/history.html" -o "$APP_DIR/frontend/history.html"
curl -sSL "$REPO/frontend/settings.html" -o "$APP_DIR/frontend/settings.html"

# 3. 安装 Go（如未安装）
echo "[3/5] 检查 Go 环境..."
if ! command -v go &>/dev/null; then
    echo "  安装 Go..."
    wget -q https://go.dev/dl/go1.21.13.linux-amd64.tar.gz -O /tmp/go.tar.gz
    tar -C /usr/local -xzf /tmp/go.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
fi

# 4. 编译
echo "[4/5] 编译程序..."
cd "$APP_DIR"
GOPROXY=https://goproxy.cn,direct go build -o okx-monitor .

# 5. 创建 systemd 服务
echo "[5/5] 创建系统服务..."
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
echo "    systemctl start $SERVICE_NAME   # 启动"
echo "    systemctl stop $SERVICE_NAME    # 停止"
echo "    systemctl restart $SERVICE_NAME # 重启"
echo "    systemctl status $SERVICE_NAME  # 状态"
echo "    journalctl -u $SERVICE_NAME -f  # 日志"
echo "========================================"
