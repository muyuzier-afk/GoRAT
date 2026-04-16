package controllers

import (
	"net/http"
	"strconv"

	"campus-management-server/models"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 简单的认证实现，实际项目中应该使用JWT
		token := c.GetHeader("Authorization")
		if token != "admin-token" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetClients 获取所有客户端
func GetClients(c *gin.Context) {
	var clients []models.Client
	models.DB.Find(&clients)

	c.JSON(http.StatusOK, gin.H{
		"clients": clients,
	})
}

// GetClientDetail 获取客户端详情
func GetClientDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	var client models.Client
	if err := models.DB.First(&client, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	// 获取客户端的文件和遥测数据
	var files []models.File
	models.DB.Where("client_id = ?", client.ID).Find(&files)

	var telemetry []models.Telemetry
	models.DB.Where("client_id = ?", client.ID).Order("created_at DESC").Limit(10).Find(&telemetry)

	c.JSON(http.StatusOK, gin.H{
		"client":    client,
		"files":     files,
		"telemetry": telemetry,
	})
}

// SendCommand 发送命令到客户端
func SendCommand(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	var req struct {
		Type    string `json:"type"`
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建命令
	command := models.Command{
		ClientID: uint(id),
		Type:     req.Type,
		Content:  req.Content,
		Status:   "pending",
	}

	models.DB.Create(&command)

	c.JSON(http.StatusOK, gin.H{
		"message": "Command sent",
		"command": command,
	})
}

// GetFiles 获取所有文件
func GetFiles(c *gin.Context) {
	var files []models.File
	models.DB.Preload("Client").Find(&files)

	c.JSON(http.StatusOK, gin.H{
		"files": files,
	})
}

// GetFileDetail 获取文件详情
func GetFileDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	var file models.File
	if err := models.DB.Preload("Client").First(&file, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"file": file,
	})
}

// DeleteFile 删除文件
func DeleteFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	// 删除文件记录
	if err := models.DB.Delete(&models.File{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully",
	})
}

// PowerControl 电源控制
func PowerControl(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	var req struct {
		Action string `json:"action"` // shutdown, restart, sleep
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建电源命令
	command := models.Command{
		ClientID: uint(id),
		Type:     "power",
		Content:  req.Action,
		Status:   "pending",
	}

	models.DB.Create(&command)

	c.JSON(http.StatusOK, gin.H{
		"message": "Power command sent",
		"command": command,
	})
}

// ProcessControl 进程控制
func ProcessControl(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client ID"})
		return
	}

	var req struct {
		Action string `json:"action"` // start, stop, kill
		Process string `json:"process"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建进程命令
	command := models.Command{
		ClientID: uint(id),
		Type:     "process",
		Content:  req.Action + " " + req.Process,
		Status:   "pending",
	}

	models.DB.Create(&command)

	c.JSON(http.StatusOK, gin.H{
		"message": "Process command sent",
		"command": command,
	})
}
