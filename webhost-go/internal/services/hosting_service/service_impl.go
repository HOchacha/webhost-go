package hosting_service

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
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

type VMStatus struct {
	VMName string
	Active bool
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

	existing, err := s.repo.FindActiveByUserID(userID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("사용자에게 이미 할당된 VM이 있습니다: %s", existing.VMName)
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("VM 존재 여부 확인 실패: %w", err)
	}

	usedIPsStr, _ := s.repo.GetUsedIPs()
	var used []net.IP
	for _, ipStr := range usedIPsStr {
		if ip := net.ParseIP(ipStr); ip != nil {
			used = append(used, ip)
		}
	}

	// 사용 가능한 IP 및 포트 확보
	ipList, err := s.Libvirt.GetUsableIPs(used)
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

func (s *HostingService) DeleteVM(email string) error {
	hostname := removeDomain(email) + "_VM"
	// 1. DB에서 VM 정보 조회
	hosting, err := s.repo.FindByVMName(hostname)
	if err != nil {
		return fmt.Errorf("VM 정보 조회 실패: %w", err)
	}

	// 2. libvirt에서 VM 삭제
	if err := s.Libvirt.DeleteDomain(hostname, true); err != nil {
		return fmt.Errorf("libvirt 도메인 삭제 실패: %w", err)
	}

	// 3. nginx-agent에 설정 제거 요청
	if err := RemoveFromNginxAgent(s.agentAddr, hosting.VMName); err != nil {
		return fmt.Errorf("nginx-agent 설정 제거 실패: %w", err)
	}

	// 4. DB에서 상태를 'deleted'로 업데이트
	if err := s.repo.UpdateStatus(hostname, "deleted"); err != nil {
		return fmt.Errorf("DB 상태 업데이트 실패: %w", err)
	}

	return nil
}

func (s *HostingService) GetVMStatus(email string) (*VMStatus, error) {
	hostname := removeDomain(email) + "_VM"
	active, err := s.Libvirt.DomainIsActive(hostname)
	if err != nil {
		return nil, fmt.Errorf("VM 상태 조회 실패: %w", err)
	}

	Status := &VMStatus{
		VMName: hostname,
		Active: active,
	}

	return Status, nil
}

func (s *HostingService) GetVMDetail(email string) (*Hosting, *libvirt.DomainInfo, error) {
	hostname := removeDomain(email) + "_VM"
	// 1. DB에서 Hosting 정보 조회
	h, err := s.repo.FindByVMName(hostname)
	if err != nil {
		return nil, nil, fmt.Errorf("DB 조회 실패: %w", err)
	}

	info, err := s.Libvirt.GetDomainInfoByName(hostname)
	if err != nil {
		return nil, nil, fmt.Errorf("도메인 정보 조회 실패: %w", err)
	}

	return h, info, nil
}

func (s *HostingService) StartVM(email string) error {
	hostname := removeDomain(email) + "_VM"
	// 1. Resume
	if err := s.Libvirt.Resume(hostname); err != nil {
		return err
	}
	// 2. DB 상태 업데이트
	if err := s.repo.UpdateStatus(hostname, "running"); err != nil {
		return fmt.Errorf("상태 갱신 실패: %w", err)
	}
	return nil
}

func (s *HostingService) StopVM(email string) error {
	hostname := removeDomain(email) + "_VM"
	// 1. Shutdown
	if err := s.Libvirt.Shutdown(hostname); err != nil {
		return err
	}
	// 2. DB 상태 업데이트
	if err := s.repo.UpdateStatus(hostname, "stopped"); err != nil {
		return fmt.Errorf("상태 갱신 실패: %w", err)
	}
	return nil
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

func RemoveFromNginxAgent(agentAddr, hostname string) error {
	req, err := http.NewRequest(http.MethodDelete,
		fmt.Sprintf("http://%s/api/nginx/%s", agentAddr, hostname), nil)
	if err != nil {
		return fmt.Errorf("요청 생성 실패: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("nginx-agent 요청 실패: %w", err)
	}
	defer resp.Body.Close()

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
