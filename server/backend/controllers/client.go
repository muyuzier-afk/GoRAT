package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"gorat-server/models"
	"gorat-server/utils"

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
			ClientID:      clientID,
			ClientKey:     clientKey,
			DeviceID:      req.DeviceID,
			Name:          req.Name,
			IP:            req.IP,
			OS:            req.OS,
			Status:        "online",
			LastHeartbeat: time.Now(),
		}
		models.DB.Create(&client)
	} else {
		// 更新客户端信息
		models.DB.Model(&client).Updates(map[string]interface{}{
			"name":           req.Name,
			"ip":             req.IP,
			"os":             req.OS,
			"status":         "online",
			"last_heartbeat": time.Now(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Client registered successfully",
		"client_id":  client.ClientID,
		"client_key": client.ClientKey,
		"client":     client,
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

	var client models.Client
	result := models.DB.Where("client_id = ? AND client_key = ?", req.ClientID, req.ClientKey).First(&client)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid client credentials"})
		return
	}

	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		s3Bucket = "gorat-data"
	}

	videoPrefix := fmt.Sprintf("video/%s/", client.DeviceID)
	infoKey := fmt.Sprintf("info/%s.json", client.DeviceID)

	videoUploadURL, err := utils.GeneratePresignedPutURL(s3Bucket, videoPrefix, 15*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate upload URL"})
		return
	}

	infoUploadURL, err := utils.GeneratePresignedPutURL(s3Bucket, infoKey, 15*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate upload URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "Config retrieved successfully",
		"s3_bucket":         s3Bucket,
		"video_upload_url":  videoUploadURL,
		"video_upload_key":  videoPrefix,
		"info_upload_url":   infoUploadURL,
		"info_upload_key":   infoKey,
		"upload_expires_in": 900,
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
		"status":         "online",
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

	var client models.Client
	if err := models.DB.Where("device_id = ?", deviceID).First(&client).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()

	fileData, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file data"})
		return
	}

	date := time.Now().Format("2006-01-02")
	s3Key := fmt.Sprintf("video/%s/%s/%s", deviceID, date, file.Filename)

	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		s3Bucket = "gorat-data"
	}
	if err := utils.UploadToS3(s3Bucket, s3Key, fileData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload to S3: %v", err)})
		return
	}

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
		DeviceID  string                   `json:"device_id"`
		CPU       float64                  `json:"cpu"`
		Memory    float64                  `json:"mem_used"`
		Processes []map[string]interface{} `json:"processes"`
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
		DeviceID  string `json:"device_id"`
		SessionID string `json:"session_id"`
		StartTime string `json:"start_time"`
		OS        string `json:"os"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var client models.Client
	if err := models.DB.Where("device_id = ?", req.DeviceID).First(&client).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	infoData, err := json.Marshal(map[string]interface{}{
		"device_id":  req.DeviceID,
		"session_id": req.SessionID,
		"start_time": req.StartTime,
		"os":         req.OS,
		"updated_at": time.Now().Format(time.RFC3339),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal info data"})
		return
	}

	s3Key := fmt.Sprintf("info/%s.json", req.DeviceID)

	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		s3Bucket = "gorat-data"
	}
	if err := utils.UploadToS3(s3Bucket, s3Key, infoData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload to S3: %v", err)})
		return
	}

	fileRecord := models.File{
		ClientID:  client.ID,
		Filename:  fmt.Sprintf("%s.json", req.DeviceID),
		Path:      s3Key,
		Size:      int64(len(infoData)),
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
