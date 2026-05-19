package models

import "time"

// Alert 异动记录
type Alert struct {
	ID          string    `json:"id"`
	Symbol      string    `json:"symbol"`       // 合约名称 e.g. BTC-USDT-SWAP
	PriceBefore float64   `json:"price_before"`  // 异动前价格
	PriceAfter  float64   `json:"price_after"`   // 异动后价格
	ChangePct   float64   `json:"change_pct"`    // 涨跌幅百分比
	Direction   string    `json:"direction"`     // up / down
	Volume      float64   `json:"volume"`        // 成交量
	Time        time.Time `json:"time"`          // 异动时间
	Opened      bool      `json:"opened"`        // 是否已开仓
}

// Position 持仓
type Position struct {
	ID          string    `json:"id"`
	AlertID     string    `json:"alert_id"`      // 关联异动ID
	Symbol      string    `json:"symbol"`
	Direction   string    `json:"direction"`     // long / short
	EntryPrice  float64   `json:"entry_price"`
	Quantity    float64   `json:"quantity"`
	OpenTime    time.Time `json:"open_time"`
	CloseTime   *time.Time `json:"close_time,omitempty"`
	ClosePrice  float64   `json:"close_price"`
	Pnl         float64   `json:"pnl"`           // 盈亏
	PnlPct      float64   `json:"pnl_pct"`       // 盈亏百分比
	Status      string    `json:"status"`        // open / closed
}

// TradeRecord 交易记录
type TradeRecord struct {
	ID         string    `json:"id"`
	Symbol     string    `json:"symbol"`
	Side       string    `json:"side"`       // buy / sell
	PositionSide string  `json:"position_side"` // long / short
	Price      float64   `json:"price"`
	Quantity   float64   `json:"quantity"`
	Time       time.Time `json:"time"`
	Remark     string    `json:"remark"`     // 备注
}

// Stats 统计数据
type Stats struct {
	TotalAlerts   int     `json:"total_alerts"`   // 总异动次数
	OpenPositions int     `json:"open_positions"` // 当前持仓数
	TotalTrades   int     `json:"total_trades"`   // 总交易次数
	WinRate       float64 `json:"win_rate"`       // 胜率
	TotalPnl      float64 `json:"total_pnl"`      // 累计盈亏
	TodayAlerts   int     `json:"today_alerts"`   // 今日异动
	TodayPnl      float64 `json:"today_pnl"`      // 今日盈亏
}

// Ticker 合约行情
type Ticker struct {
	Symbol    string  `json:"symbol"`
	LastPrice float64 `json:"last_price"`
	ChangePct float64 `json:"change_pct"` // 24h涨跌幅
	Volume24h float64 `json:"volume_24h"`
	High24h   float64 `json:"high_24h"`
	Low24h    float64 `json:"low_24h"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token"`
}

// APIResponse 通用响应
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
