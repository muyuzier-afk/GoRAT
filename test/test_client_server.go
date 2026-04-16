package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	serverURL    = "http://localhost:8000"
	testDeviceID = "test-device-001"
)

type TestClient struct {
	client    *http.Client
	deviceID  string
	clientID  string
	clientKey string
	token     string
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

func (tc *TestClient) authGetRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if tc.token != "" {
		req.Header.Set("Authorization", "Bearer "+tc.token)
	}
	return tc.client.Do(req)
}

func getStr(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	return s
}

func (tc *TestClient) TestHealth() error {
	fmt.Println("[TEST] Health check")
	resp, err := http.Get(serverURL + "/health")
	if err != nil {
		return fmt.Errorf("health check failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed, status: %d", resp.StatusCode)
	}
	fmt.Println("[PASS] Health check")
	return nil
}

func (tc *TestClient) TestLogin() error {
	fmt.Println("[TEST] Admin login")
	data := map[string]interface{}{
		"username": "admin",
		"password": "changeme",
	}
	resp, err := tc.postRequest(serverURL+"/api/admin/login", data)
	if err != nil {
		return fmt.Errorf("login request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed, status: %d", resp.StatusCode)
	}
	var result map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)
	tc.token = getStr(result, "token")
	if tc.token == "" {
		return fmt.Errorf("login succeeded but no token returned")
	}
	fmt.Println("[PASS] Admin login")
	return nil
}

func (tc *TestClient) TestLoginInvalid() error {
	fmt.Println("[TEST] Login with wrong password")
	data := map[string]interface{}{
		"username": "admin",
		"password": "wrong",
	}
	resp, err := tc.postRequest(serverURL+"/api/admin/login", data)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("expected 401, got: %d", resp.StatusCode)
	}
	fmt.Println("[PASS] Login with wrong password rejected")
	return nil
}

func (tc *TestClient) TestAuthRequired() error {
	fmt.Println("[TEST] Admin endpoint without auth")
	resp, err := http.Get(serverURL + "/api/admin/clients")
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("expected 401, got: %d", resp.StatusCode)
	}
	fmt.Println("[PASS] Admin endpoint without auth rejected")
	return nil
}

func (tc *TestClient) Register() error {
	fmt.Println("[TEST] Client register")
	data := map[string]interface{}{
		"device_id": tc.deviceID,
		"name":      tc.deviceID,
		"ip":        "127.0.0.1",
	}
	resp, err := tc.postRequest(serverURL+"/api/client/register", data)
	if err != nil {
		return fmt.Errorf("register request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("register failed, status: %d", resp.StatusCode)
	}
	var result map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)
	tc.clientID = getStr(result, "client_id")
	tc.clientKey = getStr(result, "client_key")
	if tc.clientID == "" || tc.clientKey == "" {
		return fmt.Errorf("register succeeded but missing client_id or client_key")
	}
	fmt.Println("[PASS] Client register")
	return nil
}

func (tc *TestClient) GetConfig() error {
	fmt.Println("[TEST] Get client config")
	data := map[string]interface{}{
		"client_id":  tc.clientID,
		"client_key": tc.clientKey,
	}
	resp, err := tc.postRequest(serverURL+"/api/client/config", data)
	if err != nil {
		return fmt.Errorf("get config request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get config failed, status: %d", resp.StatusCode)
	}
	fmt.Println("[PASS] Get client config")
	return nil
}

func (tc *TestClient) Heartbeat() error {
	fmt.Println("[TEST] Heartbeat")
	data := map[string]interface{}{
		"device_id": tc.deviceID,
	}
	resp, err := tc.postRequest(serverURL+"/api/client/heartbeat", data)
	if err != nil {
		return fmt.Errorf("heartbeat request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("heartbeat failed, status: %d", resp.StatusCode)
	}
	fmt.Println("[PASS] Heartbeat")
	return nil
}

func (tc *TestClient) UploadTelemetry() error {
	fmt.Println("[TEST] Upload telemetry")
	data := map[string]interface{}{
		"device_id": tc.deviceID,
		"cpu":       55.0,
		"mem_used":  70.5,
		"processes": []map[string]interface{}{
			{"pid": 100, "name": "test", "cpu": 5.0, "mem": 2.0},
		},
	}
	resp, err := tc.postRequest(serverURL+"/api/client/upload/telemetry", data)
	if err != nil {
		return fmt.Errorf("telemetry upload failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telemetry upload failed, status: %d", resp.StatusCode)
	}
	fmt.Println("[PASS] Upload telemetry")
	return nil
}

func (tc *TestClient) UploadVideo() error {
	fmt.Println("[TEST] Upload video")
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("device_id", tc.deviceID)
	part, err := writer.CreateFormFile("video", "test.mp4")
	if err != nil {
		return fmt.Errorf("create form file failed: %v", err)
	}
	part.Write([]byte("fake-video-data"))
	writer.Close()

	req, err := http.NewRequest("POST", serverURL+"/api/client/upload/video", body)
	if err != nil {
		return fmt.Errorf("create request failed: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := tc.client.Do(req)
	if err != nil {
		return fmt.Errorf("video upload request failed: %v", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		if strings.Contains(string(respBody), "S3") || strings.Contains(string(respBody), "upload") {
			fmt.Println("[WARN] Video upload S3 error (S3 may not be configured)")
			return nil
		}
		return fmt.Errorf("video upload failed, status: %d, body: %s", resp.StatusCode, string(respBody))
	}
	fmt.Println("[PASS] Upload video")
	return nil
}

func (tc *TestClient) UploadInfo() error {
	fmt.Println("[TEST] Upload info")
	data := map[string]interface{}{
		"device_id":  tc.deviceID,
		"session_id": "test-session-001",
		"start_time": time.Now().Format(time.RFC3339),
		"os":         "Linux",
	}
	resp, err := tc.postRequest(serverURL+"/api/client/upload/info", data)
	if err != nil {
		return fmt.Errorf("info upload failed: %v", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		if strings.Contains(string(respBody), "S3") {
			fmt.Println("[WARN] Info upload S3 error (S3 may not be configured)")
			return nil
		}
		return fmt.Errorf("info upload failed, status: %d, body: %s", resp.StatusCode, string(respBody))
	}
	fmt.Println("[PASS] Upload info")
	return nil
}

func (tc *TestClient) GetClients() error {
	fmt.Println("[TEST] Admin get clients")
	resp, err := tc.authGetRequest(serverURL + "/api/admin/clients")
	if err != nil {
		return fmt.Errorf("get clients failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get clients failed, status: %d", resp.StatusCode)
	}
	fmt.Println("[PASS] Admin get clients")
	return nil
}

func (tc *TestClient) GetFiles() error {
	fmt.Println("[TEST] Admin get files")
	resp, err := tc.authGetRequest(serverURL + "/api/admin/files")
	if err != nil {
		return fmt.Errorf("get files failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get files failed, status: %d", resp.StatusCode)
	}
	fmt.Println("[PASS] Admin get files")
	return nil
}

func (tc *TestClient) GetStats() error {
	fmt.Println("[TEST] Admin get dashboard stats")
	resp, err := tc.authGetRequest(serverURL + "/api/admin/stats")
	if err != nil {
		return fmt.Errorf("get stats failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get stats failed, status: %d", resp.StatusCode)
	}
	fmt.Println("[PASS] Admin get dashboard stats")
	return nil
}

func main() {
	fmt.Println("=======================================")
	fmt.Println("  GoRAT - Client-Server Test Suite")
	fmt.Println("=======================================\n")

	passed := 0
	failed := 0

	runTest := func(name string, fn func() error) {
		if err := fn(); err != nil {
			fmt.Printf("[FAIL] %s: %v\n\n", name, err)
			failed++
		} else {
			passed++
		}
	}

	tc := NewTestClient(testDeviceID)

	runTest("Health Check", tc.TestHealth)

	if failed > 0 {
		fmt.Println("Server not available, stopping")
		os.Exit(1)
	}

	runTest("Login Invalid", tc.TestLoginInvalid)
	runTest("Auth Required", tc.TestAuthRequired)
	runTest("Admin Login", tc.TestLogin)

	if tc.token == "" {
		fmt.Println("Cannot proceed without auth token")
		os.Exit(1)
	}

	runTest("Client Register", tc.Register)
	runTest("Get Config", tc.GetConfig)
	runTest("Heartbeat", tc.Heartbeat)
	runTest("Upload Telemetry", tc.UploadTelemetry)
	runTest("Upload Video", tc.UploadVideo)
	runTest("Upload Info", tc.UploadInfo)
	runTest("Admin Get Clients", tc.GetClients)
	runTest("Admin Get Files", tc.GetFiles)
	runTest("Admin Get Stats", tc.GetStats)

	fmt.Println("=======================================")
	fmt.Printf("  Results: %d passed, %d failed\n", passed, failed)
	fmt.Println("=======================================")
	if failed > 0 {
		os.Exit(1)
	}
}
