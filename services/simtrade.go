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

// SimPosition 模拟持仓
type SimPosition struct {
	ID          string    `json:"id"`
	AlertID     string    `json:"alert_id"`
	Symbol      string    `json:"symbol"`
	Direction   string    `json:"direction"` // long / short
	EntryPrice  float64   `json:"entry_price"`
	EntryTime   time.Time `json:"entry_time"`
	ClosePrice  float64   `json:"close_price"`
	CloseTime   *time.Time `json:"close_time"`
	PnlPct      float64   `json:"pnl_pct"`
	TakeProfit  float64   `json:"take_profit"`  // 止盈价
	StopLoss    float64   `json:"stop_loss"`    // 止损价
	ChangePct   float64   `json:"change_pct"`   // 触发异动的涨跌幅
	CurrPrice   float64   `json:"curr_price"`   // 当前实时价（仅持仓）
	Status      string    `json:"status"`       // open / closed
}

// SimStats 模拟统计
type SimStats struct {
	TotalTrades int     `json:"total_trades"`
	Wins        int     `json:"wins"`
	Losses      int     `json:"losses"`
	WinRate     float64 `json:"win_rate"`
	TotalPnlPct float64 `json:"total_pnl_pct"` // 累计盈亏%
}

var (
	simPositions   []*SimPosition
	simHistory     []*SimPosition
	simMu          sync.Mutex
	simRunning     bool
)

// SimAlertCh 异动信号触发后通知模拟引擎
var SimAlertCh = make(chan models.Alert, 100)

// StartSimTrade 启动模拟交易引擎
func StartSimTrade() {
	if simRunning {
		return
	}
	simRunning = true
	go simLoop()
	log.Printf("[SimTrade] 模拟交易引擎已启动")
}

// StopSimTrade 停止
func StopSimTrade() {
	simRunning = false
}

// ClearSimData 清除所有模拟数据
func ClearSimData() {
	simMu.Lock()
	defer simMu.Unlock()
	simPositions = nil
	simHistory = nil
}

// GetSimPositions 获取当前持仓（带实时价）
func GetSimPositions() []*SimPosition {
	simMu.Lock()
	defer simMu.Unlock()
	result := make([]*SimPosition, len(simPositions))
	for i, p := range simPositions {
		cp := *p
		if tp := GetPrice(p.Symbol); tp != nil {
			cp.CurrPrice = tp.LastPrice
		}
		result[i] = &cp
	}
	return result
}

// GetSimHistory 获取历史记录
func GetSimHistory() []*SimPosition {
	simMu.Lock()
	defer simMu.Unlock()
	result := make([]*SimPosition, len(simHistory))
	copy(result, simHistory)
	return result
}

// GetSimStats 获取统计
func GetSimStats() SimStats {
	simMu.Lock()
	defer simMu.Unlock()

	stats := SimStats{}
	for _, p := range simHistory {
		if p.Status == "closed" {
			stats.TotalTrades++
			stats.TotalPnlPct += p.PnlPct
			if p.PnlPct > 0 {
				stats.Wins++
			} else {
				stats.Losses++
			}
		}
	}
	if stats.TotalTrades > 0 {
		stats.WinRate = float64(stats.Wins) / float64(stats.TotalTrades) * 100
	}
	return stats
}

func simLoop() {
	for simRunning {
		select {
		case alert := <-SimAlertCh:
			openSimPosition(alert)
		default:
			checkSimPositions()
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// openSimPosition 根据异动信号开模拟仓
func openSimPosition(alert models.Alert) {
	cfg := config.Get()
	if !cfg.Settings.AutoTrade {
		return // 自动交易关闭时不模拟
	}

	simMu.Lock()
	defer simMu.Unlock()

	// 检查最大持仓限制
	maxPos := cfg.Settings.MaxPositions
	if maxPos > 0 && len(simPositions) >= maxPos {
		return
	}

	direction := "long"
	if alert.Direction == "down" {
		direction = "short"
	}

	entryPrice := alert.PriceAfter
	changePctAbs := abs(alert.ChangePct)

	var tp, sl float64
	if direction == "long" {
		tp = entryPrice * (1 + changePctAbs/100)
		sl = entryPrice * (1 - changePctAbs/100)
	} else {
		tp = entryPrice * (1 - changePctAbs/100)
		sl = entryPrice * (1 + changePctAbs/100)
	}

	pos := &SimPosition{
		ID:         fmt.Sprintf("sim%d%04d", time.Now().UnixMilli(), rand.Intn(10000)),
		AlertID:    alert.ID,
		Symbol:     alert.Symbol,
		Direction:  direction,
		EntryPrice: entryPrice,
		EntryTime:  time.Now(),
		TakeProfit: tp,
		StopLoss:   sl,
		ChangePct:  alert.ChangePct,
		Status:     "open",
	}

	simPositions = append(simPositions, pos)
	log.Printf("[SimTrade] 开仓 %s %s 入场=%.6f 止盈=%.6f 止损=%.6f",
		pos.Symbol, pos.Direction, entryPrice, tp, sl)
}

// checkSimPositions 检查持仓是否触及止盈止损
func checkSimPositions() {
	simMu.Lock()
	defer simMu.Unlock()

	var remaining []*SimPosition
	for _, pos := range simPositions {
		if pos.Status != "open" {
			continue
		}

		current := GetPrice(pos.Symbol)
		if current == nil {
			remaining = append(remaining, pos)
			continue
		}

		lastPrice := current.LastPrice
		closed := false

		if pos.Direction == "long" {
			if lastPrice >= pos.TakeProfit {
				pos.ClosePrice = pos.TakeProfit
				closed = true
			} else if lastPrice <= pos.StopLoss {
				pos.ClosePrice = pos.StopLoss
				closed = true
			}
		} else {
			if lastPrice <= pos.TakeProfit {
				pos.ClosePrice = pos.TakeProfit
				closed = true
			} else if lastPrice >= pos.StopLoss {
				pos.ClosePrice = pos.StopLoss
				closed = true
			}
		}

		if closed {
			now := time.Now()
			pos.CloseTime = &now
			pos.Status = "closed"
			pos.PnlPct = (pos.ClosePrice - pos.EntryPrice) / pos.EntryPrice * 100
			if pos.Direction == "short" {
				pos.PnlPct = -pos.PnlPct
			}
			simHistory = append(simHistory, pos)
			log.Printf("[SimTrade] 平仓 %s %s PnL=%.2f%% (TP/SL=%.6f)",
				pos.Symbol, pos.Direction, pos.PnlPct, pos.ClosePrice)
		} else {
			remaining = append(remaining, pos)
		}
	}
	simPositions = remaining
}
