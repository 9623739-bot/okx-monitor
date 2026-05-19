package services

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"okx-monitor/config"
	"okx-monitor/models"
)

// 异动通道，SSE 推送用
var (
	AlertCh      = make(chan models.Alert, 1000)
	alertStoreMu sync.Mutex
	alertStore   []models.Alert
	SnapshotPrices = make(map[string]float64) // 上次快照价格
	snapshotMu  sync.Mutex
	running     bool
)

// GetAlertStore 获取异动列表
func GetAlertStore() []models.Alert {
	alertStoreMu.Lock()
	defer alertStoreMu.Unlock()
	result := make([]models.Alert, len(alertStore))
	copy(result, alertStore)
	return result
}

// AddAlert 添加异动（供外部调用）
func AddAlert(alert models.Alert) {
	alertStoreMu.Lock()
	// 保持最多 1000 条
	if len(alertStore) >= 1000 {
		alertStore = alertStore[500:]
	}
	alertStore = append(alertStore, alert)
	alertStoreMu.Unlock()

	// 推送到 SSE 通道
	select {
	case AlertCh <- alert:
	default:
		// 通道满则丢弃，避免阻塞
	}
}

// StartMonitor 启动异动监控引擎
func StartMonitor() {
	if running {
		return
	}
	running = true
	go monitorLoop()
	log.Printf("[Monitor] 异动监控引擎已启动")
}

// StopMonitor 停止监控
func StopMonitor() {
	running = false
}

func monitorLoop() {
	for running {
		cfg := config.Get()
		interval := cfg.Settings.MonitorInterval
		if interval < 1 {
			interval = 1
		}

		// 检测异动
		detectAnomalies(cfg.Settings.MonitorThreshold)

		// 等待下一个间隔
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func detectAnomalies(threshold float64) {
	prices := GetAllPrices()
	if len(prices) == 0 {
		return
	}

	snapshotMu.Lock()
	defer snapshotMu.Unlock()

	for symbol, tp := range prices {
		lastPrice := tp.LastPrice
		snapshotPrice, exists := SnapshotPrices[symbol]

		// 首次记录，不检测异动
		if !exists {
			SnapshotPrices[symbol] = lastPrice
			continue
		}

		// 计算涨跌幅
		if snapshotPrice == 0 {
			SnapshotPrices[symbol] = lastPrice
			continue
		}

		changePct := (lastPrice - snapshotPrice) / snapshotPrice * 100

		// 更新快照（使用当前价格作为下次比较基准）
		SnapshotPrices[symbol] = lastPrice

		// 未超过阈值，忽略
		if threshold > 0 && abs(changePct) < threshold {
			continue
		}

		// 产生异动记录
		direction := "down"
		if changePct > 0 {
			direction = "up"
		}

		alert := models.Alert{
			ID:          fmt.Sprintf("%d%04d", time.Now().UnixMilli(), rand.Intn(10000)),
			Symbol:      symbol,
			PriceBefore: snapshotPrice,
			PriceAfter:  lastPrice,
			ChangePct:   changePct,
			Direction:   direction,
			Volume:      tp.Vol24h,
			Time:        time.Now(),
			Opened:      false,
		}

		AddAlert(alert)
		log.Printf("[Alert] %s %s 涨跌: %.2f%% (%.4f -> %.4f)", symbol, direction, changePct, snapshotPrice, lastPrice)
	}
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
