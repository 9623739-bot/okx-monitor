package handlers

import (
	"encoding/json"
	"fmt"
	"io"

	"okx-monitor/services"

	"github.com/gin-gonic/gin"
)

// SSEAlerts 推送异动流 (Server-Sent Events)
func SSEAlerts(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	clientGone := c.Request.Context().Done()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone:
			return false
		case alert := <-services.AlertCh:
			data, err := json.Marshal(alert)
			if err != nil {
				return true
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			return true
		}
	})
}
