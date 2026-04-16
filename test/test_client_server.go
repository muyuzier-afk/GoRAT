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
	testDeviceID = "test-device-001"
)

type TestClient struct {
	client     *http.Client
	deviceID   string
	clientID   string
	clientKey  string
}

func NewTestClient(deviceID string) *TestClient {
	return &TestClient{
		client:   &http.Client{Timeout: 30 * time.Second},
		deviceID: deviceID,
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
	fmt.Println("🔧 测试1: 客户端注册")
	fmt.Println("=======================")

	url := serverURL + "/api/client/register"
	data := map[string]interface{}{
		"device_id": tc.deviceID,
		"name":      tc.deviceID,
		"ip":        "127.0.0.1",
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

	fmt.Println("✅ 注册成功!")
	fmt.Printf("   ClientID: %s\n", tc.clientID)
	fmt.Printf("   ClientKey: %s\n\n", tc.clientKey)
	return nil
}

func (tc *TestClient) GetConfig() error {
	fmt.Println("🔧 测试2: 获取配置")
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

	fmt.Println("✅ 获取配置成功!\n")
	return nil
}

func (tc *TestClient) Heartbeat() error {
	fmt.Println("🔧 测试3: 心跳检测")
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

	fmt.Println("✅ 心跳成功!\n")
	return nil
}

func (tc *TestClient) TestHealth() error {
	fmt.Println("🔧 测试0: 健康检查")
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
	fmt.Println("  GoRAT - 测试")
	fmt.Println("=======================================\n")

	testClient := NewTestClient(testDeviceID)

	// 测试1: 健康检查
	if err := testClient.TestHealth(); err != nil {
		fmt.Printf("❌ 测试失败: %v\n", err)
		fmt.Println("请确保服务端已启动在 http://localhost:8000")
		os.Exit(1)
	}

	// 测试2: 客户端注册
	if err := testClient.Register(); err != nil {
		fmt.Printf("❌ 测试失败: %v\n", err)
		os.Exit(1)
	}

	// 测试3: 获取配置
	if err := testClient.GetConfig(); err != nil {
		fmt.Printf("❌ 测试失败: %v\n", err)
		os.Exit(1)
	}

	// 测试4: 心跳检测
	if err := testClient.Heartbeat(); err != nil {
		fmt.Printf("❌ 测试失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=======================================")
	fmt.Println("  ✅ 所有测试通过!")
	fmt.Println("=======================================")
}
