package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，实际项目中应该限制
	},
}

// WebSocketHandler WebSocket处理
func WebSocketHandler(c *gin.Context) {
	clientID := c.Param("clientId")

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
