package controllers

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gorat-server/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "change-me-in-production"
	}
	return []byte(secret)
}

func GenerateToken(username, role string) (string, error) {
	claims := Claims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "gorat-server",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format, expected Bearer token"})
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return getJWTSecret(), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminUser := os.Getenv("ADMIN_USERNAME")
	if adminUser == "" {
		adminUser = "admin"
	}
	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminPass == "" {
		adminPass = "changeme"
	}

	if req.Username != adminUser || req.Password != adminPass {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := GenerateToken(req.Username, "admin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      token,
		"expires_in": 86400,
	})
}

// GetDashboardStats 获取仪表盘统计
func GetDashboardStats(c *gin.Context) {
	var totalClients int64
	var onlineClients int64
	var totalFiles int64

	models.DB.Model(&models.Client{}).Count(&totalClients)
	models.DB.Model(&models.Client{}).Where("status = ?", "online").Count(&onlineClients)
	models.DB.Model(&models.File{}).Count(&totalFiles)

	c.JSON(http.StatusOK, gin.H{
		"total_clients":  totalClients,
		"online_clients": onlineClients,
		"total_files":    totalFiles,
	})
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
		Action  string `json:"action"` // start, stop, kill
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
