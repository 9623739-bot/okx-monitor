package main

import (
	"log"

	"okx-monitor/config"
	"okx-monitor/handlers"
	"okx-monitor/services"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Get()

	// 启动 OKX WebSocket 行情
	services.StartOKXWebSocket()

	// 启动异动监控引擎
	services.StartMonitor()

	r := gin.Default()

	// 静态文件
	r.Static("/css", "./frontend/css")
	r.Static("/js", "./frontend/js")
	r.StaticFile("/", "./frontend/index.html")
	r.StaticFile("/index.html", "./frontend/index.html")
	r.StaticFile("/dashboard.html", "./frontend/dashboard.html")
	r.StaticFile("/monitor.html", "./frontend/monitor.html")
	r.StaticFile("/history.html", "./frontend/history.html")
	r.StaticFile("/settings.html", "./frontend/settings.html")

	// API路由（部分接口需要登录）
	api := r.Group("/api")
	{
		api.POST("/login", handlers.Login)
	}

	// 需要认证的API
	auth := r.Group("/api", handlers.AuthMiddleware())
	{
		auth.GET("/alerts", handlers.GetAlerts)
		auth.GET("/alerts/stream", handlers.SSEAlerts) // SSE 实时推送
		auth.GET("/positions", handlers.GetPositions)
		auth.GET("/trades", handlers.GetTrades)
		auth.GET("/tickers", handlers.GetTickers)
		auth.GET("/stats", handlers.GetStats)
		auth.GET("/settings", handlers.GetSettings)
		auth.POST("/settings", handlers.UpdateSettings)
		auth.POST("/settings/okx-keys", handlers.UpdateOKXKeys)
		auth.GET("/logs/errors", handlers.GetErrorLogs)
	}

	log.Printf("OKX Monitor 启动在端口 %s", cfg.Port)
	log.Printf("WebSocket: 正在连接 OKX 公共频道...")
	log.Printf("监控引擎: 间隔 %d 秒, 阈值 %.1f%%", cfg.Settings.MonitorInterval, cfg.Settings.MonitorThreshold)

	r.Run(":" + cfg.Port)
}
