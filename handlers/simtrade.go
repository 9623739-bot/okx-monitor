package handlers

import (
	"net/http"

	"okx-monitor/models"
	"okx-monitor/services"

	"github.com/gin-gonic/gin"
)

// GetSimPositions 获取模拟持仓
func GetSimPositions(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "success",
		Data:    services.GetSimPositions(),
	})
}

// GetSimHistory 获取模拟历史
func GetSimHistory(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "success",
		Data:    services.GetSimHistory(),
	})
}

// GetSimStats 获取模拟统计
func GetSimStats(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "success",
		Data:    services.GetSimStats(),
	})
}
