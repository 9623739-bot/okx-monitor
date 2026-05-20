package handlers

import (
	"net/http"

	"okx-monitor/config"
	"okx-monitor/models"
	"okx-monitor/services"

	"github.com/gin-gonic/gin"
)

// GetSettings 获取设置
func GetSettings(c *gin.Context) {
	cfg := config.Get()
	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "success",
		Data:    cfg,
	})
}

// UpdateSettings 更新设置
func UpdateSettings(c *gin.Context) {
	var req config.Settings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.APIResponse{Code: 400, Message: "参数错误"})
		return
	}

	cfg := config.Get()
	cfg.Settings = req
	if err := config.Save(); err != nil {
		c.JSON(http.StatusOK, models.APIResponse{Code: 500, Message: "保存失败"})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "保存成功",
		Data:    cfg,
	})
}

// UpdateOKXKeys 更新OKX API密钥
func UpdateOKXKeys(c *gin.Context) {
	var req struct {
		ApiKey  string `json:"api_key"`
		Secret  string `json:"secret"`
		Pass    string `json:"pass"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.APIResponse{Code: 400, Message: "参数错误"})
		return
	}

	cfg := config.Get()
	cfg.OKXApiKey = req.ApiKey
	cfg.OKXSecret = req.Secret
	cfg.OKXPass = req.Pass
	config.Save()

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "API密钥更新成功",
	})
}

// GetErrorLogs 获取错误日志
func GetErrorLogs(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "success",
		Data:    services.GetErrors(),
	})
}

// ResetAll 重置异动和模拟记录
func ResetAll(c *gin.Context) {
	services.ClearAlerts()
	services.ClearSimData()
	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "已清除所有异动和模拟记录",
	})
}
