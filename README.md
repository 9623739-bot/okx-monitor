# ⚡ OKX Monitor

合约异动监控交易系统 — 实时监控 OKX 全市场永续合约价格异动，SSE 实时推送，支持顺势追单交易。

## 特性

- **实时行情** — WebSocket 接入 OKX 公共频道，毫秒级价格更新
- **异动检测** — 自定义监控间隔和涨跌幅阈值，自动捕获异动
- **SSE 推送** — Server-Sent Events 实时推送异动记录到前端
- **合约过滤** — 自动排除美股合约（instCategory=3），仅监控约 265 个加密币永续
- **浅色系 UI** — 5 页面分页设计：登录、仪表盘、异动监控、交易记录、系统设置
- **趋势追单** — 异动触发后顺势开仓（暴涨做多、暴跌做空）

## 技术栈

| 层 | 技术 |
|---|------|
| 后端 | Go + Gin |
| 数据源 | OKX WebSocket v5 |
| 实时推送 | SSE (Server-Sent Events) |
| 前端 | 原生 HTML/CSS/JS，浅色系 |
| 部署 | systemd 服务，一键脚本 |

## 页面

```
登录 → 仪表盘（统计总览）
       ├── 异动监控（实时异动列表、筛选、分页）
       ├── 交易记录（持仓历史、盈亏统计）
       └── 系统设置（阈值、间隔、API密钥）
```

## 快速部署

```bash
curl -sSL https://raw.githubusercontent.com/9623739-bot/okx-monitor/master/deploy.sh | bash
```

默认端口 8088，登录 `admin` / `admin123`。

## 本地开发

```bash
git clone git@github.com:9623739-bot/okx-monitor.git
cd okx-monitor
GOPROXY=https://goproxy.cn,direct go mod tidy
go run .
```

浏览器打开 `http://localhost:8088`。

## 配置说明

| 参数 | 默认值 | 说明 |
|------|--------|------|
| 监控间隔 | 1 秒 | 1-60 秒可调 |
| 异动阈值 | 1% | 涨跌幅超过此值触发异动 |
| 最大持仓 | 3 | 同时持有的最大合约数 |
| 单笔金额 | 100 USDT | 每次开仓保证金 |
| 自动交易 | 关闭 | 开启后异动触发自动下单 |

配置通过 Web 设置页保存到 `config.json`。

## 项目结构

```
okx-monitor/
├── main.go              # 入口
├── config/
│   └── config.go        # 配置管理
├── handlers/
│   ├── auth.go          # JWT 认证
│   ├── monitor.go       # 异动/行情 API
│   ├── settings.go      # 设置 API
│   └── sse.go           # SSE 推送
├── models/
│   └── models.go        # 数据模型
├── services/
│   ├── okx_ws.go        # OKX WebSocket + 合约白名单
│   └── monitor.go       # 异动检测引擎
├── frontend/
│   ├── index.html       # 登录
│   ├── dashboard.html   # 仪表盘
│   ├── monitor.html     # 异动监控
│   ├── history.html     # 交易记录
│   ├── settings.html    # 系统设置
│   ├── css/base.css     # 浅色系主题
│   └── js/common.js     # 公共工具
└── deploy.sh            # 一键部署脚本
```

## License

MIT
