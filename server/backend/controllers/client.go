package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"campus-management-server/models"
	"campus-management-server/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RegisterClient 客户端注册
func RegisterClient(c *gin.Context) {
	var req struct {
		DeviceID string `json:"device_id"`
		Name     string `json:"name"`
		IP       string `json:"ip"`
		OS       string `json:"os"` // Windows, Linux
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找或创建客户端
	var client models.Client
	result := models.DB.Where("device_id = ?", req.DeviceID).First(&client)
	
	clientID := uuid.New().String()
	clientKey := uuid.New().String()
	
	if result.Error != nil {
		// 创建新客户端
		client = models.Client{
			ClientID:     clientID,
			ClientKey:    clientKey,
			DeviceID:     req.DeviceID,
			Name:         req.Name,
			IP:           req.IP,
			OS:           req.OS,
			Status:       "online",
			LastHeartbeat: time.Now(),
		}
		models.DB.Create(&client)
	} else {
		// 更新客户端信息
		models.DB.Model(&client).Updates(map[string]interface{}{
			"name":          req.Name,
			"ip":            req.IP,
			"os":            req.OS,
			"status":        "online",
			"last_heartbeat": time.Now(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Client registered successfully",
		"client_id": client.ClientID,
		"client_key": client.ClientKey,
		"client":  client,
	})
}

// GetClientConfig 获取客户端配置
func GetClientConfig(c *gin.Context) {
	var req struct {
		ClientID  string `json:"client_id"`
		ClientKey string `json:"client_key"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证客户端
	var client models.Client
	result := models.DB.Where("client_id = ? AND client_key = ?", req.ClientID, req.ClientKey).First(&client)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid client credentials"})
		return
	}

	// 返回S3配置
	config := gin.H{
		"s3_endpoint":  os.Getenv("S3_ENDPOINT"),
		"s3_access_key": os.Getenv("S3_ACCESS_KEY"),
		"s3_secret_key": os.Getenv("S3_SECRET_KEY"),
		"s3_region":    os.Getenv("S3_REGION"),
		"s3_bucket":     os.Getenv("S3_BUCKET"),
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Config retrieved successfully",
		"config": config,
	})
}

// Heartbeat 客户端心跳
func Heartbeat(c *gin.Context) {
	var req struct {
		DeviceID string `json:"device_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新心跳时间
	result := models.DB.Model(&models.Client{}).Where("device_id = ?", req.DeviceID).Updates(map[string]interface{}{
		"status":        "online",
		"last_heartbeat": time.Now(),
	})

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Heartbeat received"})
}

// UploadVideo 上传视频文件
func UploadVideo(c *gin.Context) {
	deviceID := c.PostForm("device_id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id is required"})
		return
	}

	// 获取客户端
	var client models.Client
	if err := models.DB.Where("device_id = ?", deviceID).First(&client).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	// 处理文件上传
	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 读取文件内容
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()

	// 构建S3路径
	date := time.Now().Format("2006-01-02")
	timeStr := time.Now().Format("15_04_05")
	s3Key := fmt.Sprintf("video/%s/%s/%s", deviceID, date, file.Filename)

	// 上传到S3
	// 注意：这里需要实现文件内容的读取和上传

	// 保存文件记录
	fileRecord := models.File{
		ClientID:  client.ID,
		Filename:  file.Filename,
		Path:      s3Key,
		Size:      file.Size,
		Type:      "video",
		S3Key:     s3Key,
		CreatedAt: time.Now(),
	}
	models.DB.Create(&fileRecord)

	c.JSON(http.StatusOK, gin.H{
		"message": "Video uploaded successfully",
		"file":    fileRecord,
	})
}

// UploadTelemetry 上传遥测数据
func UploadTelemetry(c *gin.Context) {
	var req struct {
		DeviceID   string          `json:"device_id"`
		CPU        float64         `json:"cpu"`
		Memory     float64         `json:"mem_used"`
		Processes  []map[string]interface{} `json:"processes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取客户端
	var client models.Client
	if err := models.DB.Where("device_id = ?", req.DeviceID).First(&client).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	// 序列化进程列表
	processesJSON, _ := json.Marshal(req.Processes)

	// 保存遥测数据
	telemetry := models.Telemetry{
		ClientID:  client.ID,
		CPU:       req.CPU,
		Memory:    req.Memory,
		Processes: string(processesJSON),
		CreatedAt: time.Now(),
	}
	models.DB.Create(&telemetry)

	c.JSON(http.StatusOK, gin.H{
		"message":   "Telemetry uploaded successfully",
		"telemetry": telemetry,
	})
}

// UploadInfo 上传设备信息
func UploadInfo(c *gin.Context) {
	var req struct {
		DeviceID   string `json:"device_id"`
		SessionID  string `json:"session_id"`
		StartTime  string `json:"start_time"`
		OS         string `json:"os"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取客户端
	var client models.Client
	if err := models.DB.Where("device_id = ?", req.DeviceID).First(&client).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	// 构建S3路径
	s3Key := fmt.Sprintf("info/%s.json", req.DeviceID)

	// 上传到S3
	// 注意：这里需要实现数据的上传

	// 保存文件记录
	fileRecord := models.File{
		ClientID:  client.ID,
		Filename:  fmt.Sprintf("%s.json", req.DeviceID),
		Path:      s3Key,
		Size:      int64(len([]byte(req.OS))),
		Type:      "info",
		S3Key:     s3Key,
		CreatedAt: time.Now(),
	}
	models.DB.Create(&fileRecord)

	c.JSON(http.StatusOK, gin.H{
		"message": "Info uploaded successfully",
		"file":    fileRecord,
	})
}
