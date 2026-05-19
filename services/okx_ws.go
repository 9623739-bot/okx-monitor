package services

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// OKXTicker OKX WebSocket ticker 数据
type OKXTicker struct {
	InstType string `json:"instType"`
	InstID   string `json:"instId"`
	Last     string `json:"last"`
	Open24h  string `json:"open24h"`
	High24h  string `json:"high24h"`
	Low24h   string `json:"low24h"`
	Vol24h   string `json:"vol24h"`
}

// OKXWsMsg OKX WebSocket 消息
type OKXWsMsg struct {
	Event string      `json:"event"`
	Arg   OKXWsArg    `json:"arg"`
	Data  []OKXTicker `json:"data"`
}

type OKXWsArg struct {
	Channel  string `json:"channel"`
	InstID   string `json:"instId"`
	InstType string `json:"instType"`
}

// OKXInstrument 产品信息
type OKXInstrument struct {
	InstType     string `json:"instType"`
	InstID       string `json:"instId"`
	InstFamily   string `json:"instFamily"`
	InstCategory string `json:"instCategory"`
}

type OKXInstResponse struct {
	Code string           `json:"code"`
	Data []OKXInstrument  `json:"data"`
}

// TickerPrice 内存中的价格缓存
type TickerPrice struct {
	Symbol    string
	LastPrice float64
	Open24h   float64
	High24h   float64
	Low24h    float64
	Vol24h    float64
	UpdatedAt time.Time
}

// PriceStore 价格存储
var (
	priceMap  = make(map[string]*TickerPrice)
	priceMu   sync.RWMutex
	allowSet  = make(map[string]bool) // 允许的 instId 白名单
	allowMu   sync.RWMutex
	wsRunning bool
	wsStop    chan struct{}
)

// GetPrice 获取合约当前价格
func GetPrice(symbol string) *TickerPrice {
	priceMu.RLock()
	defer priceMu.RUnlock()
	return priceMap[symbol]
}

// GetAllPrices 获取所有合约价格
func GetAllPrices() map[string]*TickerPrice {
	priceMu.RLock()
	defer priceMu.RUnlock()
	result := make(map[string]*TickerPrice, len(priceMap))
	for k, v := range priceMap {
		result[k] = v
	}
	return result
}

// GetPriceMap 获取价格原始map
func GetPriceMap() map[string]*TickerPrice {
	return priceMap
}

// PriceMapMutex 获取价格锁
func PriceMapMutex() *sync.RWMutex {
	return &priceMu
}

// fetchInstruments 获取产品列表并建立白名单（排除美股合约）
func fetchInstruments() int {
	resp, err := http.Get("https://www.okx.com/api/v5/public/instruments?instType=SWAP")
	if err != nil {
		log.Printf("[INST] 获取产品列表失败: %v", err)
		return 0
	}
	defer resp.Body.Close()

	var result OKXInstResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[INST] 解析产品列表失败: %v", err)
		return 0
	}

	if result.Code != "0" {
		log.Printf("[INST] API错误: code=%s", result.Code)
		return 0
	}

	allowMu.Lock()
	defer allowMu.Unlock()

	cryptoCount := 0
	excludedCount := 0
	allowSet = make(map[string]bool)

	for _, inst := range result.Data {
		// 排除美股合约 instCategory=3 和期权
		if inst.InstCategory == "3" {
			excludedCount++
			continue
		}
		// 只保留加密币永续
		if inst.InstCategory == "1" && isUSDTPerpetual(inst.InstID) {
			allowSet[inst.InstID] = true
			cryptoCount++
		}
	}

	log.Printf("[INST] 产品白名单: %d 个加密币永续, 已排除 %d 个美股/期权合约", cryptoCount, excludedCount)
	return cryptoCount
}

// StartOKXWebSocket 启动 OKX WebSocket 连接
func StartOKXWebSocket() {
	// 先获取产品白名单
	count := fetchInstruments()
	if count == 0 {
		log.Printf("[WS] 未能获取产品列表，跳过启动")
		return
	}

	wsStop = make(chan struct{})
	go wsLoop()
}

// StopOKXWebSocket 停止
func StopOKXWebSocket() {
	if wsStop != nil {
		close(wsStop)
		wsRunning = false
	}
}

func wsLoop() {
	url := "wss://ws.okx.com:8443/ws/v5/public"
	wsRunning = true

	for wsRunning {
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Printf("[WS] 连接失败: %v, 5秒后重试", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Printf("[WS] 已连接 OKX 公共频道")

		// 订阅所有永续合约 tickers
		subscribe := map[string]interface{}{
			"op": "subscribe",
			"args": []map[string]string{
				{"channel": "tickers", "instType": "SWAP"},
			},
		}
		if err := conn.WriteJSON(subscribe); err != nil {
			log.Printf("[WS] 订阅失败: %v", err)
			conn.Close()
			time.Sleep(5 * time.Second)
			continue
		}
		log.Printf("[WS] 已订阅所有永续合约 tickers")

		// 读取消息循环
		readLoop(conn)
		conn.Close()
		log.Printf("[WS] 连接断开，5秒后重连")
		time.Sleep(5 * time.Second)
	}
}

func readLoop(conn *websocket.Conn) {
	for {
		select {
		case <-wsStop:
			return
		default:
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[WS] 读取错误: %v", err)
			return
		}

		var wsMsg OKXWsMsg
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			continue
		}

		// 只处理 tickers 数据
		if wsMsg.Arg.Channel != "tickers" || len(wsMsg.Data) == 0 {
			continue
		}

		for _, ticker := range wsMsg.Data {
			// 白名单过滤：排除美股等非加密币合约
			allowMu.RLock()
			allowed := allowSet[ticker.InstID]
			allowMu.RUnlock()
			if !allowed {
				continue
			}

			last := parseFloat(ticker.Last)
			if last == 0 {
				continue
			}

			priceMu.Lock()
			priceMap[ticker.InstID] = &TickerPrice{
				Symbol:    ticker.InstID,
				LastPrice: last,
				Open24h:   parseFloat(ticker.Open24h),
				High24h:   parseFloat(ticker.High24h),
				Low24h:    parseFloat(ticker.Low24h),
				Vol24h:    parseFloat(ticker.Vol24h),
				UpdatedAt: time.Now(),
			}
			priceMu.Unlock()
		}
	}
}

// isUSDTPerpetual 判断是否为 USDT 本位永续合约
func isUSDTPerpetual(instID string) bool {
	if len(instID) < 10 {
		return false
	}
	if instID[len(instID)-4:] != "SWAP" {
		return false
	}
	for i := 0; i <= len(instID)-4; i++ {
		if instID[i:i+4] == "USDT" {
			return true
		}
	}
	return false
}

func parseFloat(s string) float64 {
	var f float64
	json.Unmarshal([]byte(s), &f)
	return f
}
