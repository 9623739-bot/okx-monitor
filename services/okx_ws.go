package services

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// OKXTicker 行情数据
type OKXTicker struct {
	InstType string `json:"instType"`
	InstID   string `json:"instId"`
	Last     string `json:"last"`
	Open24h  string `json:"open24h"`
	High24h  string `json:"high24h"`
	Low24h   string `json:"low24h"`
	Vol24h   string `json:"vol24h"`
}

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
	Code string          `json:"code"`
	Data []OKXInstrument `json:"data"`
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

var (
	priceMap  = make(map[string]*TickerPrice)
	priceMu   sync.RWMutex
	allowSet  = make(map[string]bool)
	allowMu   sync.RWMutex
)

// GetPrice / GetAllPrices 等接口保持不变
func GetPrice(symbol string) *TickerPrice {
	priceMu.RLock()
	defer priceMu.RUnlock()
	return priceMap[symbol]
}

func GetAllPrices() map[string]*TickerPrice {
	priceMu.RLock()
	defer priceMu.RUnlock()
	r := make(map[string]*TickerPrice, len(priceMap))
	for k, v := range priceMap {
		r[k] = v
	}
	return r
}

func GetPriceMap() map[string]*TickerPrice          { return priceMap }
func PriceMapMutex() *sync.RWMutex                   { return &priceMu }

// fetchInstruments 获取产品白名单（排除美股）
func fetchInstruments() int {
	resp, err := http.Get("https://www.okx.com/api/v5/public/instruments?instType=SWAP")
	if err != nil {
		log.Printf("[INST] 获取产品列表失败: %v", err)
		return 0
	}
	defer resp.Body.Close()

	var result OKXInstResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[INST] 解析失败: %v", err)
		return 0
	}
	if result.Code != "0" {
		log.Printf("[INST] API错误 code=%s", result.Code)
		return 0
	}

	allowMu.Lock()
	defer allowMu.Unlock()

	cryptoCount := 0
	excludedCount := 0
	allowSet = make(map[string]bool)
	for _, inst := range result.Data {
		if inst.InstCategory == "3" {
			excludedCount++
			continue
		}
		if inst.InstCategory == "1" && isUSDT(inst.InstID) {
			allowSet[inst.InstID] = true
			cryptoCount++
		}
	}
	log.Printf("[INST] 白名单: %d 加密币合约, 排除 %d 美股/期权", cryptoCount, excludedCount)
	return cryptoCount
}

// StartOKXWebSocket 启动行情
func StartOKXWebSocket() {
	fetchInstruments()
	go wsLoop()
}

// StopOKXWebSocket 停止
func StopOKXWebSocket() {}

func wsLoop() {
	url := "wss://ws.okx.com:8443/ws/v5/public"
	log.Printf("[WS] 初始化连接...")

	// 构建 instId 列表
	allowMu.RLock()
	ids := make([]string, 0, len(allowSet))
	for id := range allowSet {
		ids = append(ids, id)
	}
	allowMu.RUnlock()
	log.Printf("[WS] 准备订阅 %d 个合约", len(ids))

	for {
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Printf("[WS] 连接失败: %v, 5s重试", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Printf("[WS] 已连接")

		// 分批订阅（每批100个）
		batch := 100
		for i := 0; i < len(ids); i += batch {
			end := i + batch
			if end > len(ids) {
				end = len(ids)
			}
			args := make([]map[string]string, end-i)
			for j, id := range ids[i:end] {
				args[j] = map[string]string{"channel": "tickers", "instId": id}
			}
			sub := map[string]interface{}{"op": "subscribe", "args": args}
			if err := conn.WriteJSON(sub); err != nil {
				log.Printf("[WS] 订阅失败: %v", err)
				break
			}
		}
		log.Printf("[WS] 订阅已发送")

		// 读消息
		tickerCount := 0
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("[WS] 断开 (%d条ticker): %v", tickerCount, err)
				break
			}

			var wsMsg OKXWsMsg
			if err := json.Unmarshal(msg, &wsMsg); err != nil {
				var raw struct{ Event string; Code string; Msg string }
				if json.Unmarshal(msg, &raw) == nil {
					if raw.Event != "" {
						log.Printf("[WS] 事件: %s code=%s msg=%s", raw.Event, raw.Code, raw.Msg)
					}
				}
				continue
			}

			if wsMsg.Arg.Channel != "tickers" || len(wsMsg.Data) == 0 {
				continue
			}

			tickerCount++
			priceMu.Lock()
			for _, t := range wsMsg.Data {
				allowMu.RLock()
				if !allowSet[t.InstID] {
					allowMu.RUnlock()
					continue
				}
				allowMu.RUnlock()
				last := parseFloat(t.Last)
				if last == 0 {
					continue
				}
				priceMap[t.InstID] = &TickerPrice{
					Symbol: t.InstID, LastPrice: last,
					Open24h: parseFloat(t.Open24h), High24h: parseFloat(t.High24h),
					Low24h: parseFloat(t.Low24h), Vol24h: parseFloat(t.Vol24h),
					UpdatedAt: time.Now(),
				}
			}
			priceMu.Unlock()
		}
		conn.Close()
		log.Printf("[WS] 5s后重连...")
		time.Sleep(5 * time.Second)
	}
}

func isUSDT(instID string) bool {
	if len(instID) < 11 {
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
