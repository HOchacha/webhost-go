package nginx_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNginxManager(t *testing.T) {
	tmpDir := t.TempDir()

	// 테스트 경로 세팅
	hostingFile := filepath.Join(tmpDir, "sites-available", "web_host.conf")
	streamDir := filepath.Join(tmpDir, "stream.d")

	// 디렉터리 생성
	os.MkdirAll(filepath.Dir(hostingFile), 0755)
	os.MkdirAll(streamDir, 0755)

	manager := NewNginxManager(hostingFile, streamDir)

	// Agent 정보
	agent := AgentInfo{
		Username: "testuser",
		Hostname: "testuser",
		VMIP:     "192.168.100.100",
	}

	// nginx 템플릿 정의
	nginxConfTemplate = `
# BEGIN WEBHOSTING_Hochacha {{.Hostname}}
location /code/{{.Username}}/ {
    proxy_pass http://{{.VMIP}}:8080/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}
# END WEBHOSTING_Hochacha {{.Hostname}}`
	streamConfTemplate = `
server {
    listen 22022;
    proxy_pass {{.VMIP}}:22;
}`

	// 1. HTTP Config 추가
	if err := manager.AddHTTPConfig(agent); err != nil {
		t.Fatalf("AddHTTPConfig failed: %v", err)
	}

	// 2. Stream Config 추가
	if err := manager.AddStreamConfig(agent); err != nil {
		t.Fatalf("AddStreamConfig failed: %v", err)
	}

	// 3. 삭제
	if err := manager.RemoveNginxConfigForUser(agent.Username); err != nil {
		t.Fatalf("RemoveNginxConfigForUser failed: %v", err)
	}
}
