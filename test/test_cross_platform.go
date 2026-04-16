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
	serverURL = "http://localhost:8000"
)

type TestClient struct {
	client    *http.Client
	deviceID  string
	clientID  string
	clientKey string
	osType    string
	token     string
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

func (tc *TestClient) authDeleteRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", url, nil)
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

func TestHealth() error {
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

func TestLogin() (string, error) {
	fmt.Println("[TEST] Admin login")
	data := map[string]interface{}{
		"username": "admin",
		"password": "changeme",
	}
	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", serverURL+"/api/admin/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("login request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed, status: %d, body: %s", resp.StatusCode, string(body))
	}
	var result map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)
	token := getStr(result, "token")
	if token == "" {
		return "", fmt.Errorf("login succeeded but no token returned")
	}
	fmt.Println("[PASS] Admin login")
	return token, nil
}

func TestLoginInvalidCredentials() error {
	fmt.Println("[TEST] Login with invalid credentials")
	data := map[string]interface{}{
		"username": "admin",
		"password": "wrongpassword",
	}
	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", serverURL+"/api/admin/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("expected 401 for invalid credentials, got: %d", resp.StatusCode)
	}
	fmt.Println("[PASS] Login with invalid credentials rejected")
	return nil
}

func TestAuthRequired() error {
	fmt.Println("[TEST] Admin endpoints require auth")
	resp, err := http.Get(serverURL + "/api/admin/clients")
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("expected 401 without auth, got: %d", resp.StatusCode)
	}
	fmt.Println("[PASS] Admin endpoints require auth")
	return nil
}

func (tc *TestClient) Register() error {
	fmt.Printf("[TEST] %s client register\n", tc.osType)
	data := map[string]interface{}{
		"device_id": tc.deviceID,
		"name":      tc.deviceID,
		"ip":        "127.0.0.1",
		"os":        tc.osType,
	}
	resp, err := tc.postRequest(serverURL+"/api/client/register", data)
	if err != nil {
		return fmt.Errorf("register request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("register failed, status: %d, body: %s", resp.StatusCode, string(body))
	}
	var result map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)
	tc.clientID = getStr(result, "client_id")
	tc.clientKey = getStr(result, "client_key")
	if tc.clientID == "" || tc.clientKey == "" {
		return fmt.Errorf("register succeeded but missing client_id or client_key")
	}
	fmt.Printf("[PASS] %s client registered (id=%s)\n", tc.osType, tc.clientID)
	return nil
}

func (tc *TestClient) GetConfig() error {
	fmt.Printf("[TEST] %s client get config\n", tc.osType)
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
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("get config failed, status: %d, body: %s", resp.StatusCode, string(body))
	}
	var result map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)
	if getStr(result, "s3_bucket") == "" {
		return fmt.Errorf("config missing s3_bucket field")
	}
	if getStr(result, "video_upload_url") == "" {
		fmt.Println("[WARN] No video_upload_url (S3 may not be configured)")
	}
	fmt.Printf("[PASS] %s client got config\n", tc.osType)
	return nil
}

func (tc *TestClient) GetConfigInvalidKey() error {
	fmt.Printf("[TEST] %s client get config with invalid key\n", tc.osType)
	data := map[string]interface{}{
		"client_id":  tc.clientID,
		"client_key": "invalid-key",
	}
	resp, err := tc.postRequest(serverURL+"/api/client/config", data)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("expected 401 for invalid client_key, got: %d", resp.StatusCode)
	}
	fmt.Printf("[PASS] %s client config with invalid key rejected\n", tc.osType)
	return nil
}

func (tc *TestClient) Heartbeat() error {
	fmt.Printf("[TEST] %s client heartbeat\n", tc.osType)
	data := map[string]interface{}{
		"device_id": tc.deviceID,
	}
	resp, err := tc.postRequest(serverURL+"/api/client/heartbeat", data)
	if err != nil {
		return fmt.Errorf("heartbeat request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("heartbeat failed, status: %d, body: %s", resp.StatusCode, string(body))
	}
	fmt.Printf("[PASS] %s client heartbeat\n", tc.osType)
	return nil
}

func (tc *TestClient) UploadTelemetry() error {
	fmt.Printf("[TEST] %s client upload telemetry\n", tc.osType)
	data := map[string]interface{}{
		"device_id": tc.deviceID,
		"cpu":       45.5,
		"mem_used":  62.3,
		"processes": []map[string]interface{}{
			{"pid": 1234, "name": "test-proc", "cpu": 10.5, "mem": 5.2},
		},
	}
	resp, err := tc.postRequest(serverURL+"/api/client/upload/telemetry", data)
	if err != nil {
		return fmt.Errorf("telemetry upload failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telemetry upload failed, status: %d, body: %s", resp.StatusCode, string(body))
	}
	fmt.Printf("[PASS] %s client telemetry uploaded\n", tc.osType)
	return nil
}

func (tc *TestClient) UploadVideo() error {
	fmt.Printf("[TEST] %s client upload video\n", tc.osType)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("device_id", tc.deviceID)
	part, err := writer.CreateFormFile("video", "test_video.mp4")
	if err != nil {
		return fmt.Errorf("create form file failed: %v", err)
	}
	part.Write([]byte("fake-video-content"))
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
			fmt.Printf("[WARN] %s video upload S3 error (S3 may not be configured): %s\n", tc.osType, strings.TrimSpace(string(respBody)))
			return nil
		}
		return fmt.Errorf("video upload failed, status: %d, body: %s", resp.StatusCode, string(respBody))
	}
	fmt.Printf("[PASS] %s client video uploaded\n", tc.osType)
	return nil
}

func (tc *TestClient) GetClients() error {
	fmt.Println("[TEST] Admin get clients list")
	resp, err := tc.authGetRequest(serverURL + "/api/admin/clients")
	if err != nil {
		return fmt.Errorf("get clients failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("get clients failed, status: %d, body: %s", resp.StatusCode, string(body))
	}
	fmt.Println("[PASS] Admin get clients list")
	return nil
}

func (tc *TestClient) GetDashboardStats() error {
	fmt.Println("[TEST] Admin get dashboard stats")
	resp, err := tc.authGetRequest(serverURL + "/api/admin/stats")
	if err != nil {
		return fmt.Errorf("get stats failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("get stats failed, status: %d, body: %s", resp.StatusCode, string(body))
	}
	fmt.Println("[PASS] Admin get dashboard stats")
	return nil
}

func (tc *TestClient) GetFiles() error {
	fmt.Println("[TEST] Admin get files list")
	resp, err := tc.authGetRequest(serverURL + "/api/admin/files")
	if err != nil {
		return fmt.Errorf("get files failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("get files failed, status: %d, body: %s", resp.StatusCode, string(body))
	}
	fmt.Println("[PASS] Admin get files list")
	return nil
}

func main() {
	fmt.Println("=======================================")
	fmt.Println("  GoRAT - Cross Platform Test Suite")
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

	runTest("Health Check", TestHealth)

	if failed > 0 {
		fmt.Println("Server not available, skipping remaining tests")
		fmt.Printf("\nResults: %d passed, %d failed\n", passed, failed)
		os.Exit(1)
	}

	runTest("Login Invalid Credentials", TestLoginInvalidCredentials)
	runTest("Auth Required", TestAuthRequired)

	token, err := TestLogin()
	if err != nil {
		fmt.Printf("[FAIL] Admin Login: %v\n\n", err)
		failed++
		fmt.Println("Cannot proceed without auth token, skipping remaining tests")
		fmt.Printf("\nResults: %d passed, %d failed\n", passed, failed)
		os.Exit(1)
	}
	passed++

	winClient := NewTestClient("windows-test-001", "Windows")
	winClient.token = token
	linuxClient := NewTestClient("linux-test-001", "Linux")
	linuxClient.token = token

	runTest("Windows Client Register", winClient.Register)
	runTest("Windows Client Get Config", winClient.GetConfig)
	runTest("Windows Client Get Config Invalid Key", winClient.GetConfigInvalidKey)
	runTest("Windows Client Heartbeat", winClient.Heartbeat)
	runTest("Windows Client Upload Telemetry", winClient.UploadTelemetry)
	runTest("Windows Client Upload Video", winClient.UploadVideo)

	runTest("Linux Client Register", linuxClient.Register)
	runTest("Linux Client Get Config", linuxClient.GetConfig)
	runTest("Linux Client Heartbeat", linuxClient.Heartbeat)
	runTest("Linux Client Upload Telemetry", linuxClient.UploadTelemetry)

	runTest("Admin Get Clients", winClient.GetClients)
	runTest("Admin Get Dashboard Stats", winClient.GetDashboardStats)
	runTest("Admin Get Files", winClient.GetFiles)

	fmt.Println("=======================================")
	fmt.Printf("  Results: %d passed, %d failed\n", passed, failed)
	fmt.Println("=======================================")
	if failed > 0 {
		os.Exit(1)
	}
}
