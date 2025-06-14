package nginx_test

import (
	"os"
	"path/filepath"
	"testing"
	"webhost-go/webhost-go/cmd/nginx-agent/nginx"
)

func TestNginxManager_HTTPAndStreamConfig(t *testing.T) {
	// ────────────────
	// 1. 임시 디렉토리 구성
	tmpDir := t.TempDir()
	locationDir := filepath.Join(tmpDir, "locations")
	streamDir := filepath.Join(tmpDir, "stream.d")

	_ = os.MkdirAll(locationDir, 0755)
	_ = os.MkdirAll(streamDir, 0755)

	hostingFile := filepath.Join(tmpDir, "webhost.conf")
	err := os.WriteFile(hostingFile, []byte(`server {
    listen 80;
    server_name _;
    include locations/*.conf;
}`), 0644)
	if err != nil {
		t.Fatalf("failed to create base config file: %v", err)
	}

	manager := nginx.NewNginxManager(hostingFile, locationDir, streamDir)

	// ────────────────
	// 2. AgentInfo 테스트 데이터
	agent := nginx.AgentInfo{
		Username: "testuser",
		Hostname: "testvm.local",
		VMIP:     "192.168.0.123",
		SSHPort:  22022,
	}

	// ────────────────
	// 3. AddHTTPConfig + AddStreamConfig 테스트
	if err := manager.AddHTTPConfig(agent); err != nil {
		t.Fatalf("AddHTTPConfig failed: %v", err)
	}

	if err := manager.AddStreamConfig(agent); err != nil {
		t.Fatalf("AddStreamConfig failed: %v", err)
	}

	locPath := filepath.Join(locationDir, "testuser.conf")
	streamPath := filepath.Join(streamDir, "sftp_testuser.conf")

	if _, err := os.Stat(locPath); os.IsNotExist(err) {
		t.Fatalf("expected HTTP config file not found: %s", locPath)
	}
	if _, err := os.Stat(streamPath); os.IsNotExist(err) {
		t.Fatalf("expected stream config file not found: %s", streamPath)
	}

	// ────────────────
	// 4. RemoveNginxConfigForUser 테스트
	if err := manager.RemoveNginxConfigForUser(agent.Username); err != nil {
		t.Fatalf("RemoveNginxConfigForUser failed: %v", err)
	}

	if _, err := os.Stat(locPath); !os.IsNotExist(err) {
		t.Errorf("HTTP config file should be deleted: %s", locPath)
	}
	if _, err := os.Stat(streamPath); !os.IsNotExist(err) {
		t.Errorf("stream config file should be deleted: %s", streamPath)
	}
}
