package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/mail"
	"net/mime"
	"net/mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// 配置
const (
	serverEndpoint = "http://localhost:8000" // 服务端地址
	deviceID       = "factory-001"             // 每台设备不同
	configFile     = "factoryeye_config.json"  // 配置文件
)

// 操作系统类型
var osType string

func init() {
	// 通过编译参数设置操作系统类型
	// 编译时使用: go build -ldflags "-X main.osType=Windows" 或 "-X main.osType=Linux"
	if osType == "" {
		// 自动检测操作系统
		if runtime.GOOS == "windows" {
			osType = "Windows"
		} else {
			osType = "Linux"
		}
	}
}

type ClientConfig struct {
	S3Endpoint  string `json:"s3_endpoint"`
	S3AccessKey string `json:"s3_access_key"`
	S3SecretKey string `json:"s3_secret_key"`
	S3Region    string `json:"s3_region"`
	S3Bucket    string `json:"s3_bucket"`
}

type AgentCredentials struct {
	ClientID  string `json:"client_id"`
	ClientKey string `json:"client_key"`
}

type Agent struct {
	serverClient  *http.Client
	sessionID     string
	credentials   *AgentCredentials
	clientConfig  *ClientConfig
}

func main() {
	// 单实例检查
	if !isSingleInstance() {
		log.Println("程序已在运行中")
		os.Exit(0)
	}

	// 初始化服务器客户端
	serverClient := initServerClient()

	agent := &Agent{
		serverClient: serverClient,
		sessionID:    generateSessionID(),
	}

	// 尝试加载或注册凭证
	if err := agent.loadOrRegisterCredentials(); err != nil {
		log.Fatalf("Failed to load/register credentials: %v", err)
	}

	// 从服务端获取S3配置
	if err := agent.fetchClientConfig(); err != nil {
		log.Fatalf("Failed to fetch client config: %v", err)
	}

	// 上传设备信息
	agent.uploadDeviceInfo()

	// 启动系统信息采集协程
	go agent.collectAndUploadSystemInfo()

	// 启动视频采集协程
	go agent.recordAndUpload()

	// 启动心跳协程
	go agent.sendHeartbeat()

	// 保持程序运行
	select {}
}

func isSingleInstance() bool {
	if osType == "Windows" {
		// 在Windows上使用tasklist命令
		cmd := exec.Command("cmd", "/c", "tasklist", "/FI", "IMAGENAME eq factoryeye.exe")
		output, err := cmd.Output()
		if err != nil {
			return true
		}
		count := strings.Count(string(output), "factoryeye.exe")
		return count <= 1
	} else {
		// 在Linux上使用ps命令
		cmd := exec.Command("ps", "aux")
		output, err := cmd.Output()
		if err != nil {
			return true
		}
		count := strings.Count(string(output), "factoryeye")
		return count <= 1
	}
}

func initServerClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}

func (a *Agent) loadOrRegisterCredentials() error {
	// 尝试从文件加载凭证
	if a.tryLoadCredentials() {
		return nil
	}

	// 注册新客户端
	return a.registerNewClient()
}

func (a *Agent) tryLoadCredentials() bool {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return false
	}

	var creds AgentCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return false
	}

	a.credentials = &creds
	return true
}

func (a *Agent) saveCredentials() error {
	data, err := json.Marshal(a.credentials)
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0600)
}

func (a *Agent) registerNewClient() error {
	url := serverEndpoint + "/api/client/register"
	data := map[string]interface{}{
		"device_id": deviceID,
		"name":      deviceID,
		"ip":        getLocalIP(),
		"os":        osType,
	}

	resp, err := a.postRequestWithResponse(url, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	clientID, _ := result["client_id"].(string)
	clientKey, _ := result["client_key"].(string)

	a.credentials = &AgentCredentials{
		ClientID:  clientID,
		ClientKey: clientKey,
	}

	return a.saveCredentials()
}

func (a *Agent) fetchClientConfig() error {
	url := serverEndpoint + "/api/client/config"
	data := map[string]interface{}{
		"client_id":  a.credentials.ClientID,
		"client_key": a.credentials.ClientKey,
	}

	resp, err := a.postRequestWithResponse(url, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	configData, _ := result["config"].(map[string]interface{})
	a.clientConfig = &ClientConfig{
		S3Endpoint:  configData["s3_endpoint"].(string),
		S3AccessKey: configData["s3_access_key"].(string),
		S3SecretKey: configData["s3_secret_key"].(string),
		S3Region:    configData["s3_region"].(string),
		S3Bucket:    configData["s3_bucket"].(string),
	}

	return nil
}

func (a *Agent) postRequestWithResponse(url string, data interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return a.serverClient.Do(req)
}

func (a *Agent) sendHeartbeat() {
	for {
		url := serverEndpoint + "/api/client/heartbeat"
		data := map[string]interface{}{
			"device_id": deviceID,
		}
		a.postRequest(url, data)
		time.Sleep(60 * time.Second)
	}
}

func (a *Agent) postRequest(url string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.serverClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func generateSessionID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func (a *Agent) uploadFile(fileType string, data []byte, filename string) error {
	url := fmt.Sprintf("http://localhost:8000/api/client/upload/%s", fileType)

	// 创建multipart表单
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	// 添加device_id字段
	w.WriteField("device_id", deviceID)

	// 添加文件字段
	fw, err := w.CreateFormFile(fileType, filename)
	if err != nil {
		return err
	}
	_, err = fw.Write(data)
	if err != nil {
		return err
	}
	w.Close()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := a.serverClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (a *Agent) uploadDeviceInfo() {
	data := map[string]interface{}{
		"device_id":  deviceID,
		"session_id": a.sessionID,
		"start_time": time.Now().Format(time.RFC3339),
		"os":         getOSInfo(),
	}

	url := serverEndpoint + "/api/client/upload/info"
	a.postRequest(url, data)
}

func getOSInfo() string {
	if osType == "Windows" {
		cmd := exec.Command("cmd", "/c", "ver")
		output, err := cmd.Output()
		if err != nil {
			return "Windows"
		}
		return strings.TrimSpace(string(output))
	} else {
		cmd := exec.Command("uname", "-a")
		output, err := cmd.Output()
		if err != nil {
			return "Linux"
		}
		return strings.TrimSpace(string(output))
	}
}

func getLocalIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip = ip.To4()
			if ip == nil {
				continue
			}

			return ip.String()
		}
	}

	return "127.0.0.1"
}

func (a *Agent) collectAndUploadSystemInfo() {
	for {
		// 采集CPU
		cpuPercent, _ := cpu.Percent(0, false)
		// 采集内存
		memInfo, _ := mem.VirtualMemory()
		// 采集进程列表
		processes, _ := process.Processes()
		var procList []map[string]interface{}

		// 限制进程数量，避免数据过大
		limit := 50
		if len(processes) < limit {
			limit = len(processes)
		}

		for _, p := range processes[:limit] {
			name, _ := p.Name()
			cpu, _ := p.CPUPercent()
			mem, _ := p.MemoryPercent()
			procList = append(procList, map[string]interface{}{
				"pid":  p.Pid,
				"name": name,
				"cpu":  cpu,
				"mem":  mem,
			})
		}

		data := map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"cpu":       cpuPercent[0],
			"mem_used":  memInfo.UsedPercent,
			"processes": procList,
		}

		url := serverEndpoint + "/api/client/upload/telemetry"
	data["device_id"] = deviceID
	a.postRequest(url, data)

		time.Sleep(30 * time.Second)
	}
}

func (a *Agent) recordAndUpload() {
	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "factoryeye")
	os.MkdirAll(tempDir, 0755)

	// 使用ffmpeg进行视频采集和编码
	// 注意：需要在系统PATH中添加ffmpeg
	segmentCounter := 1
	for {
		timestamp := time.Now().Format("2006-01-02_15_04_05")
		segmentFile := filepath.Join(tempDir, fmt.Sprintf("%s_%03d.mp4", timestamp, segmentCounter))

		var cmd *exec.Cmd
		if osType == "Windows" {
			// Windows设备
			cmd = exec.Command(
				"ffmpeg",
				"-f", "dshow",
				"-i", "video=Integrated Camera:audio=Microphone Array",
				"-c:v", "libx264",
				"-preset", "veryfast",
				"-crf", "23",
				"-c:a", "aac",
				"-b:a", "128k",
				"-t", "5",
				"-y",
				segmentFile,
			)
		} else {
			// Linux设备
			cmd = exec.Command(
				"ffmpeg",
				"-f", "v4l2",
				"-i", "/dev/video0",
				"-f", "alsa",
				"-i", "default",
				"-c:v", "libx264",
				"-preset", "veryfast",
				"-crf", "23",
				"-c:a", "aac",
				"-b:a", "128k",
				"-t", "5",
				"-y",
				segmentFile,
			)
		}

		err := cmd.Run()
		if err != nil {
			log.Printf("ffmpeg执行失败: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// 读取文件并上传
		data, err := os.ReadFile(segmentFile)
		if err == nil {
			// 上传到服务器
			a.uploadFile("video", data, filepath.Base(segmentFile))
		}

		// 删除临时文件
		os.Remove(segmentFile)

		segmentCounter++
		if segmentCounter > 999 {
			segmentCounter = 1
		}
	}
}
