package handlers

import (
	"net/http"
	"sync"
	"time"

	"okx-monitor/models"
	"okx-monitor/services"

	"github.com/gin-gonic/gin"
)

var (
	positionsStore []models.Position
	tradesStore   []models.TradeRecord
	storeMu       sync.Mutex
)

func init() {
	// 保留模拟持仓，交易功能后续实现
	now := time.Now()
	positionsStore = []models.Position{
		{ID: "p1", AlertID: "a1", Symbol: "BTC-USDT-SWAP", Direction: "long", EntryPrice: 68320.00, Quantity: 0.01, OpenTime: now.Add(-10 * time.Minute), Status: "open", Pnl: 45.50, PnlPct: 0.67},
		{ID: "p2", AlertID: "a2", Symbol: "ETH-USDT-SWAP", Direction: "short", EntryPrice: 3412.80, Quantity: 0.1, OpenTime: now.Add(-8 * time.Minute), Status: "open", Pnl: -12.30, PnlPct: -0.36},
	}
}

// GetAlerts 获取异动记录（从 service 实时数据）
func GetAlerts(c *gin.Context) {
	alerts := services.GetAlertStore()
	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "success",
		Data:    alerts,
	})
}

// GetPositions 获取持仓
func GetPositions(c *gin.Context) {
	storeMu.Lock()
	defer storeMu.Unlock()

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "success",
		Data:    positionsStore,
	})
}

// GetTrades 获取交易记录
func GetTrades(c *gin.Context) {
	storeMu.Lock()
	defer storeMu.Unlock()

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "success",
		Data:    tradesStore,
	})
}

// GetTickers 获取合约行情（从 WS 实时数据）
func GetTickers(c *gin.Context) {
	prices := services.GetAllPrices()
	tickers := make([]models.Ticker, 0, len(prices))
	for _, p := range prices {
		tickers = append(tickers, models.Ticker{
			Symbol:    p.Symbol,
			LastPrice: p.LastPrice,
			High24h:   p.High24h,
			Low24h:    p.Low24h,
			Volume24h: p.Vol24h,
		})
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "success",
		Data:    tickers,
	})
}

// GetStats 获取统计数据
func GetStats(c *gin.Context) {
	alerts := services.GetAlertStore()
	storeMu.Lock()
	defer storeMu.Unlock()

	stats := models.Stats{
		TotalAlerts:   len(alerts),
		OpenPositions: 0,
		TotalTrades:   len(tradesStore),
		TodayAlerts:   len(alerts),
	}

	for _, p := range positionsStore {
		if p.Status == "open" {
			stats.OpenPositions++
		}
		stats.TotalPnl += p.Pnl
	}

	stats.TodayPnl = stats.TotalPnl

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "success",
		Data:    stats,
	})
}
