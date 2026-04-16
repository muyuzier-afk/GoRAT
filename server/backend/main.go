package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"campus-management-server/controllers"
	"campus-management-server/models"
	"campus-management-server/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// 初始化数据库
	if err := models.InitDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 初始化S3客户端
	if err := utils.InitS3Client(); err != nil {
		log.Fatalf("Failed to initialize S3 client: %v", err)
	}

	// 启动心跳检查协程
	go models.CheckHeartbeats()

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎
	r := gin.Default()

	// 配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API路由组
	api := r.Group("/api")
	{
		// 客户端相关路由
		client := api.Group("/client")
		{
			client.POST("/register", controllers.RegisterClient)
			client.POST("/config", controllers.GetClientConfig)
			client.POST("/heartbeat", controllers.Heartbeat)
			client.POST("/upload/video", controllers.UploadVideo)
			client.POST("/upload/telemetry", controllers.UploadTelemetry)
			client.POST("/upload/info", controllers.UploadInfo)
		}

		// 管理员相关路由
		admin := api.Group("/admin")
		admin.Use(controllers.AuthMiddleware())
		{
			// 客户端管理
			admin.GET("/clients", controllers.GetClients)
			admin.GET("/client/:id", controllers.GetClientDetail)
			admin.POST("/client/:id/command", controllers.SendCommand)

			// 文件管理
			admin.GET("/files", controllers.GetFiles)
			admin.GET("/file/:id", controllers.GetFileDetail)
			admin.DELETE("/file/:id", controllers.DeleteFile)

			// 设备控制
			admin.POST("/client/:id/power", controllers.PowerControl)
			admin.POST("/client/:id/process", controllers.ProcessControl)
		}

		// WebSocket路由
		api.GET("/ws/:clientId", controllers.WebSocketHandler)
	}

	// 静态文件服务
	r.Static("/static", "./static")

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Server starting on port %s...", port)
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
