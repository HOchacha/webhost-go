package hosting_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
	"webhost-go/webhost-go/cmd/nginx-agent/nginx"
	"webhost-go/webhost-go/pkg/libvirt"
)

type HostingService struct {
	repo      HostingRepository
	agentAddr string
	Libvirt   *libvirt.LibvirtManager
}

type VMRequest struct {
	Name  string `json:"name"`
	Owner string `json:"owner"` // 사용자 ID
}

type AgentInfo struct {
	UserID    int64  `json:"user_id"`
	IPAddress string `json:"ip"`
	SSHPort   int    `json:"ssh_port"`
	ProxyPath string `json:"proxy_path"` // 예: "/123" 또는 "/vm-abc"
}

type UserVM struct {
	Name string
	IP   net.IP
	Port int
}

func NewService(repo HostingRepository, agentAddr string, libvirt *libvirt.LibvirtManager) *HostingService {
	return &HostingService{
		repo:      repo,
		agentAddr: agentAddr,
		Libvirt:   libvirt,
	}
}

// CreateHosting - 새로운 VM을 생성하고 Hosting 엔트리를 DB에 등록
func (s *HostingService) CreateHosting(userID int64, email string) (*Hosting, error) {
	username := removeDomain(email)
	hostname := removeDomain(email) + "_VM"

	// 사용 가능한 IP 및 포트 확보
	ipList, err := s.Libvirt.GetUsableIPs()
	if err != nil || len(ipList) == 0 {
		return nil, fmt.Errorf("사용 가능한 IP 없음: %w", err)
	}
	ip := ipList[0]

	port, err := s.repo.GetAvailablePort(20000, 30000)
	if err != nil {
		return nil, fmt.Errorf("사용 가능한 포트 없음: %w", err)
	}

	// VM 생성
	if err := s.Libvirt.StartUbuntuVMWithStaticIP(hostname, ip); err != nil {
		return nil, fmt.Errorf("VM 생성 실패: %w", err)
	}

	// nginx-agent 등록
	agent := nginx.AgentInfo{
		Username: username,
		Hostname: hostname,
		VMIP:     ip.String(),
		SSHPort:  port,
	}
	if err := RegisterWithNginxAgent(s.agentAddr, agent); err != nil {
		return nil, err
	}

	// DB에 기록
	h := &Hosting{
		UserID:    userID,
		VMName:    hostname,
		IPAddress: ip.String(),
		SSHPort:   port,
		ProxyPath: "/" + username,
		DiskPath:  fmt.Sprintf("/var/lib/libvirt/images/instances/%s/disk.qcow2", hostname),
		Status:    "running",
		CreatedAt: time.Now(),
	}
	if err := s.repo.Create(h); err != nil {
		return nil, fmt.Errorf("DB 저장 실패: %w", err)
	}

	return h, nil
}

func (s *HostingService) DeleteVM(name string) error {
	return s.Libvirt.DeleteDomain(name, true)
}

func (s *HostingService) GetVMStatus(name string) (bool, error) {
	// 나중에 DomainIsActive() 등 상태 조회 로직 추가 가능
	return true, nil
}

func (s *HostingService) StopVM(name string) error {
	return s.Libvirt.Shutdown(name)
}

// RegisterWithNginxAgent sends AgentInfo to nginx-agent for proxy registration
func RegisterWithNginxAgent(agentAddr string, agent nginx.AgentInfo) error {
	url := fmt.Sprintf("http://%s/api/nginx", agentAddr)

	// JSON 직렬화
	body, err := json.Marshal(agent)
	if err != nil {
		return fmt.Errorf("nginx-agent 전송 실패: JSON 변환 오류: %w", err)
	}

	// HTTP POST 요청
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("nginx-agent 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	// 응답 코드 확인
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("nginx-agent 오류 응답: %s", string(data))
	}

	return nil
}

func removeDomain(email string) string {
	if at := strings.Index(email, "@"); at != -1 {
		return email[:at]
	}
	return email // @가 없는 경우 그대로 반환
}
