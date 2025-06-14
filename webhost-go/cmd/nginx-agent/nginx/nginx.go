package nginx

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

var (
	hostingFile = "/usr/local/nginx/conf/sites-available"
	streamFile  = "/usr/local/nginx/conf/stream.d/sftp_"
)

type NginxManager struct {
	HostingFilePath string // ex: /usr/local/nginx/conf/sites-available/webhost.conf
	LocationDirPath string // ex: /usr/local/nginx/conf/sites-available/locations/
	StreamDirPath   string // ex: /usr/local/nginx/conf/stream.d/
}

func NewNginxManager(hostingFile, locationDir, streamDir string) *NginxManager {
	return &NginxManager{
		HostingFilePath: hostingFile,
		LocationDirPath: locationDir,
		StreamDirPath:   streamDir,
	}
}

func (n *NginxManager) AddHTTPConfig(agent AgentInfo) error {
	tmpl, err := template.New("http").Parse(nginxConfTemplate)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, agent); err != nil {
		return err
	}

	// /sites-available/locations/username.conf
	confPath := filepath.Join(n.LocationDirPath, fmt.Sprintf("%s.conf", agent.Username))
	if err := os.MkdirAll(n.LocationDirPath, 0755); err != nil {
		return fmt.Errorf("failed to ensure locations dir exists: %w", err)
	}
	if err := os.WriteFile(confPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write location config: %w", err)
	}

	return nil
}

func (n *NginxManager) AddStreamConfig(agent AgentInfo) error {
	tmpl, err := template.New("stream").Parse(streamConfTemplate)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, agent); err != nil {
		return err
	}

	confPath := filepath.Join(n.StreamDirPath, fmt.Sprintf("sftp_%s.conf", agent.Username))
	return os.WriteFile(confPath, buf.Bytes(), 0644)
}

func (n *NginxManager) Reload() error {
	return exec.Command("nginx", "-s", "reload").Run()
}

func (n *NginxManager) RemoveNginxConfigForUser(hostname string) error {
	// ────────────────────────────────────────────────────────
	// 1. HTTP 프록시 설정 삭제
	locationFile := filepath.Join(n.LocationDirPath, fmt.Sprintf("%s.conf", hostname))
	if err := os.Remove(locationFile); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("ℹ️  No HTTP location file to delete for user: %s\n", hostname)
		} else {
			return fmt.Errorf("Failed to remove HTTP location config %s: %v", locationFile, err)
		}
	} else {
		fmt.Printf("✅ HTTP location config removed: %s\n", locationFile)
	}

	// ────────────────────────────────────────────────────────
	// 2. Stream 설정 파일 삭제
	streamPath := fmt.Sprintf("%s/sftp_%s.conf", n.StreamDirPath, hostname)
	if err := os.Remove(streamPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("ℹ️  No stream config to delete for user: %s\n", hostname)
		} else {
			return fmt.Errorf("Failed to remove stream config file %s: %v", streamPath, err)
		}
	} else {
		fmt.Printf("✅ Stream config removed: %s\n", streamPath)
	}

	// ────────────────────────────────────────────────────────
	return nil
}

// 파일 생성 함수
func createConfigFile(filename string) error {
	defaultContent := `server {
    listen 80;
    server_name _;
}`
	// 디렉터리가 없으면 생성
	dir := "/etc/nginx/sites-available"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// 파일 생성
	err := ioutil.WriteFile(filename, []byte(defaultContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	fmt.Println("Config file created successfully.")
	return nil
}
