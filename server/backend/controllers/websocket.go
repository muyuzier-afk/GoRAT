package controllers

import (
	"log"
	"net/http"
	"os"
	"strings"

	"gorat-server/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func isOriginAllowed(origin string) bool {
	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:3000"
	}
	for _, o := range strings.Split(allowedOrigins, ",") {
		if strings.TrimSpace(o) == origin {
			return true
		}
	}
	return false
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return false
		}
		return isOriginAllowed(origin)
	},
}

// WebSocketHandler WebSocket处理
func WebSocketHandler(c *gin.Context) {
	clientID := c.Param("clientId")
	clientKey := c.GetHeader("X-Client-Key")

	// 验证客户端身份
	var client models.Client
	if err := models.DB.Where("client_id = ? AND status = ?", clientID, "online").First(&client).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or offline client"})
		return
	}

	// 验证 clientKey
	if clientKey == "" || clientKey != client.ClientKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid client key"})
		return
	}

	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// 处理WebSocket连接
	for {
		// 读取消息
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// 处理消息
		log.Printf("Received message from client %s: %s", clientID, message)

		// 发送响应
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}
}
