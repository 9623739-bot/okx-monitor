package handlers

import (
	"net/http"
	"time"

	"okx-monitor/config"
	"okx-monitor/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Login 登录
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.APIResponse{Code: 400, Message: "参数错误"})
		return
	}

	cfg := config.Get()
	if req.Username != cfg.AdminUser || req.Password != cfg.AdminPass {
		c.JSON(http.StatusOK, models.APIResponse{Code: 401, Message: "用户名或密码错误"})
		return
	}

	token, err := generateToken(req.Username, cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusOK, models.APIResponse{Code: 500, Message: "生成Token失败"})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    200,
		Message: "登录成功",
		Data:    models.LoginResponse{Token: token},
	})
}

func generateToken(username, secret string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// AuthMiddleware JWT验证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 静态资源和登录页不需要验证
		path := c.Request.URL.Path
		if path == "/" || path == "/index.html" || path == "/api/login" ||
			len(path) > 4 && path[:4] == "/css" ||
			len(path) > 3 && path[:3] == "/js" {
			c.Next()
			return
		}

		// 优先从 Header 取，其次从 query string（兼容 EventSource SSE）
		tokenStr := ""
		if auth := c.GetHeader("Authorization"); auth != "" && len(auth) > 7 && auth[:7] == "Bearer " {
			tokenStr = auth[7:]
		} else if qs := c.Query("token"); qs != "" {
			tokenStr = qs
		}

		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, models.APIResponse{Code: 401, Message: "未登录"})
			c.Abort()
			return
		}

		cfg := config.Get()
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, models.APIResponse{Code: 401, Message: "Token无效"})
			c.Abort()
			return
		}

		c.Next()
	}
}
