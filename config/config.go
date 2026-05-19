package config

import (
	"encoding/json"
	"os"
	"sync"
)

// Settings 系统设置
type Settings struct {
	MonitorInterval  int     `json:"monitor_interval"`  // 监控间隔(秒)
	MonitorThreshold float64 `json:"monitor_threshold"` // 异动阈值(百分比)
	MaxPositions     int     `json:"max_positions"`     // 最大持仓数
	OrderSize        float64 `json:"order_size"`        // 单笔开仓数量(USDT)
	AutoTrade        bool    `json:"auto_trade"`        // 是否自动交易
}

// AppConfig 应用配置
type AppConfig struct {
	Port        string   `json:"port"`
	AdminUser   string   `json:"admin_user"`
	AdminPass   string   `json:"admin_pass"`
	JWTSecret   string   `json:"jwt_secret"`
	OKXApiKey   string   `json:"okx_api_key"`
	OKXSecret   string   `json:"okx_secret"`
	OKXPass     string   `json:"okx_pass"`
	Settings    Settings `json:"settings"`
}

var (
	cfg  *AppConfig
	once sync.Once
)

// Get 获取配置单例
func Get() *AppConfig {
	once.Do(func() {
		cfg = &AppConfig{
			Port:      "8088",
			AdminUser: "admin",
			AdminPass: "admin123",
			JWTSecret: "okx-monitor-secret-key-2024",
			Settings: Settings{
				MonitorInterval:  1,
				MonitorThreshold: 1.0,
				MaxPositions:     3,
				OrderSize:        100,
				AutoTrade:        false,
			},
		}
		loadFromFile()
	})
	return cfg
}

func loadFromFile() {
	data, err := os.ReadFile("config.json")
	if err != nil {
		return
	}
	json.Unmarshal(data, cfg)
}

// Save 保存配置到文件
func Save() error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("config.json", data, 0644)
}
