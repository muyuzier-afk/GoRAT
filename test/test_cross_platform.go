package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	serverURL = "http://localhost:8000"
)

type TestClient struct {
	client     *http.Client
	deviceID   string
	clientID   string
	clientKey  string
	osType     string
}

func NewTestClient(deviceID, osType string) *TestClient {
	return &TestClient{
		client:   &http.Client{Timeout: 30 * time.Second},
		deviceID: deviceID,
		osType:   osType,
	}
}

func (tc *TestClient) postRequest(url string, data interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return tc.client.Do(req)
}

func (tc *TestClient) Register() error {
	fmt.Printf("🔧 测试: %s客户端注册\n", tc.osType)
	fmt.Println("=======================")

	url := serverURL + "/api/client/register"
	data := map[string]interface{}{
		"device_id": tc.deviceID,
		"name":      tc.deviceID,
		"ip":        "127.0.0.1",
		"os":        tc.osType,
	}

	resp, err := tc.postRequest(url, data)
	if err != nil {
		return fmt.Errorf("注册请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("注册失败，状态码: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("响应: %s\n\n", string(body))

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	tc.clientID = result["client_id"].(string)
	tc.clientKey = result["client_key"].(string)

	fmt.Printf("✅ %s客户端注册成功!\n", tc.osType)
	fmt.Printf("   ClientID: %s\n", tc.clientID)
	fmt.Printf("   ClientKey: %s\n\n", tc.clientKey)
	return nil
}

func (tc *TestClient) GetConfig() error {
	fmt.Printf("🔧 测试: %s客户端获取配置\n", tc.osType)
	fmt.Println("=======================")

	url := serverURL + "/api/client/config"
	data := map[string]interface{}{
		"client_id":  tc.clientID,
		"client_key": tc.clientKey,
	}

	resp, err := tc.postRequest(url, data)
	if err != nil {
		return fmt.Errorf("获取配置请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("获取配置失败，状态码: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("响应: %s\n\n", string(body))

	fmt.Printf("✅ %s客户端获取配置成功!\n\n", tc.osType)
	return nil
}

func (tc *TestClient) Heartbeat() error {
	fmt.Printf("🔧 测试: %s客户端心跳\n", tc.osType)
	fmt.Println("=======================")

	url := serverURL + "/api/client/heartbeat"
	data := map[string]interface{}{
		"device_id": tc.deviceID,
	}

	resp, err := tc.postRequest(url, data)
	if err != nil {
		return fmt.Errorf("心跳请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("心跳失败，状态码: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("响应: %s\n\n", string(body))

	fmt.Printf("✅ %s客户端心跳成功!\n\n", tc.osType)
	return nil
}

func TestHealth() error {
	fmt.Println("🔧 测试: 健康检查")
	fmt.Println("=======================")

	resp, err := http.Get(serverURL + "/health")
	if err != nil {
		return fmt.Errorf("健康检查失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("健康检查失败，状态码: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("响应: %s\n\n", string(body))

	fmt.Println("✅ 健康检查通过!\n")
	return nil
}

func main() {
	fmt.Println("=======================================")
	fmt.Println("  GoRAT - 跨平台测试")
	fmt.Println("=======================================\n")

	// 测试1: 健康检查
	if err := TestHealth(); err != nil {
		fmt.Printf("❌ 测试失败: %v\n", err)
		fmt.Println("请确保服务端已启动在 http://localhost:8000")
		os.Exit(1)
	}

	// 测试2: Windows客户端
	windowsClient := NewTestClient("windows-test-001", "Windows")
	if err := windowsClient.Register(); err != nil {
		fmt.Printf("❌ Windows客户端测试失败: %v\n", err)
		os.Exit(1)
	}
	if err := windowsClient.GetConfig(); err != nil {
		fmt.Printf("❌ Windows客户端测试失败: %v\n", err)
		os.Exit(1)
	}
	if err := windowsClient.Heartbeat(); err != nil {
		fmt.Printf("❌ Windows客户端测试失败: %v\n", err)
		os.Exit(1)
	}

	// 测试3: Linux客户端
	linuxClient := NewTestClient("linux-test-001", "Linux")
	if err := linuxClient.Register(); err != nil {
		fmt.Printf("❌ Linux客户端测试失败: %v\n", err)
		os.Exit(1)
	}
	if err := linuxClient.GetConfig(); err != nil {
		fmt.Printf("❌ Linux客户端测试失败: %v\n", err)
		os.Exit(1)
	}
	if err := linuxClient.Heartbeat(); err != nil {
		fmt.Printf("❌ Linux客户端测试失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=======================================")
	fmt.Println("  ✅ 所有跨平台测试通过!")
	fmt.Println("=======================================")
}
