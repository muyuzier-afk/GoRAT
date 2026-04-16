package models

import (
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Client 客户端模型
type Client struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ClientID      string    `gorm:"uniqueIndex" json:"client_id"`
	ClientKey     string    `json:"client_key"`
	DeviceID      string    `json:"device_id"`
	Name          string    `json:"name"`
	IP            string    `json:"ip"`
	OS            string    `json:"os"`     // Windows, Linux
	Status        string    `json:"status"` // online, offline
	LastHeartbeat time.Time `json:"last_heartbeat"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// 关联
	Files     []File      `json:"files,omitempty"`
	Telemetry []Telemetry `json:"telemetry,omitempty"`
	Commands  []Command   `json:"commands,omitempty"`
}

// File 文件模型
type File struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ClientID  uint      `json:"client_id"`
	Filename  string    `json:"filename"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	Type      string    `json:"type"` // video, telemetry, info
	S3Key     string    `json:"s3_key"`
	CreatedAt time.Time `json:"created_at"`

	// 关联
	Client Client `json:"client,omitempty"`
}

// Telemetry 遥测数据模型
type Telemetry struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ClientID  uint      `json:"client_id"`
	CPU       float64   `json:"cpu"`
	Memory    float64   `json:"memory"`
	Processes string    `json:"processes"` // JSON string
	CreatedAt time.Time `json:"created_at"`

	// 关联
	Client Client `json:"client,omitempty"`
}

// Command 命令模型
type Command struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ClientID  uint      `json:"client_id"`
	Type      string    `json:"type"` // power, process, shell
	Content   string    `json:"content"`
	Status    string    `json:"status"` // pending, executed, failed
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联
	Client Client `json:"client,omitempty"`
}

// InitDB 初始化数据库连接
func InitDB() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=gorat port=5432 sslmode=disable"
	}
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	err = DB.AutoMigrate(&Client{}, &File{}, &Telemetry{}, &Command{})
	if err != nil {
		return err
	}

	return nil
}

// CheckHeartbeats 检查客户端心跳
func CheckHeartbeats() {
	for {
		time.Sleep(60 * time.Second)

		// 标记超过5分钟没有心跳的客户端为离线
		offlineThreshold := time.Now().Add(-5 * time.Minute)
		DB.Model(&Client{}).Where("last_heartbeat < ?", offlineThreshold).Update("status", "offline")
	}
}
