package hosting_service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	awscli "github.com/aws/aws-sdk-go-v2/aws"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
	"webhost-go/webhost-go/cmd/nginx-agent/nginx"
	"webhost-go/webhost-go/pkg/aws"
)

type HostingService struct {
	repo      HostingRepository
	agentAddr string
	EC2       *aws.EC2Manager // 기존 libvirt 대신 EC2Manager 사용
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

func NewService(repo HostingRepository, agentAddr string, ec2mgr *aws.EC2Manager) *HostingService {
	return &HostingService{
		repo:      repo,
		agentAddr: agentAddr,
		EC2:       ec2mgr,
	}
}

// CreateHosting - 새로운 VM을 생성하고 Hosting 엔트리를 DB에 등록
func (s *HostingService) CreateHosting(userID int64, email string) (*Hosting, error) {
	username := removeDomain(email)
	hostname := username + "_VM"

	existing, err := s.repo.FindActiveByUserID(userID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("사용자에게 이미 할당된 VM이 있습니다: %s", existing.VMName)
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("VM 존재 여부 확인 실패: %w", err)
	}

	port, err := s.repo.GetAvailablePort(20000, 30000)
	if err != nil {
		return nil, fmt.Errorf("사용 가능한 포트 없음: %w", err)
	}

	ctx := context.Background()
	instanceID, _, privateIP, err := s.EC2.StartUbuntuVMWithOwner(ctx, hostname, username)
	if err != nil {
		return nil, fmt.Errorf("EC2 인스턴스 생성 실패: %w", err)
	}

	// nginx-agent 등록
	agent := nginx.AgentInfo{
		Username: username,
		Hostname: hostname,
		VMIP:     privateIP,
		SSHPort:  port,
	}
	if err := RegisterWithNginxAgent(s.agentAddr, agent); err != nil {
		return nil, err
	}

	h := &Hosting{
		UserID:    userID,
		VMName:    hostname,
		IPAddress: privateIP,
		SSHPort:   port,
		ProxyPath: "/" + username,
		DiskPath:  instanceID, // AWS에서는 디스크 경로 대신 인스턴스 ID로 추적
		Status:    "running",
		CreatedAt: time.Now(),
	}
	if err := s.repo.Create(h); err != nil {
		return nil, fmt.Errorf("DB 저장 실패: %w", err)
	}

	return h, nil
}

func (s *HostingService) DeleteVM(email string) error {
	username := removeDomain(email)
	hostname := username + "_VM"

	// 1. DB 조회
	_, err := s.repo.FindByVMName(hostname)
	if err != nil {
		return fmt.Errorf("VM 정보 조회 실패: %w", err)
	}

	// 2. EC2 종료
	if err := s.EC2.TerminateVMByUser(context.Background(), username); err != nil {
		return fmt.Errorf("EC2 인스턴스 종료 실패: %w", err)
	}

	// 3. nginx-agent 제거
	if err := RemoveFromNginxAgent(s.agentAddr, hostname); err != nil {
		return fmt.Errorf("nginx-agent 설정 제거 실패: %w", err)
	}

	// 4. DB 상태 업데이트
	if err := s.repo.UpdateStatus(hostname, "deleted"); err != nil {
		return fmt.Errorf("DB 상태 업데이트 실패: %w", err)
	}

	return nil
}

func (s *HostingService) GetVMStatus(email string) (*VMStatus, error) {
	username := removeDomain(email)
	active, err := s.EC2.GetVMStatusByUser(context.Background(), username)
	if err != nil {
		return nil, fmt.Errorf("VM 상태 조회 실패: %w", err)
	}
	return &VMStatus{
		VMName: username + "_VM",
		Active: active == "running",
	}, nil
}

type EC2InstanceInfo struct {
	InstanceID   string
	State        string
	InstanceType string
	PublicIP     string
	LaunchTime   time.Time
}

func (s *HostingService) GetVMDetail(email string) (*Hosting, *EC2InstanceInfo, error) {
	hostname := removeDomain(email) + "_VM"
	h, err := s.repo.FindByVMName(hostname)
	if err != nil {
		return nil, nil, fmt.Errorf("DB 조회 실패: %w", err)
	}

	username := removeDomain(email)
	ctx := context.Background()

	inst, err := s.EC2.FindInstanceByUser(ctx, username)
	if err != nil {
		return nil, nil, fmt.Errorf("EC2 인스턴스 조회 실패: %w", err)
	}

	info := &EC2InstanceInfo{
		InstanceID:   awscli.ToString(inst.InstanceId),
		State:        string(inst.State.Name),
		InstanceType: string(inst.InstanceType),
		PublicIP:     awscli.ToString(inst.PublicIpAddress),
		LaunchTime:   awscli.ToTime(inst.LaunchTime),
	}

	return h, info, nil
}

func (s *HostingService) StartVM(email string) error {
	username := removeDomain(email)

	if err := s.EC2.StartVMByUser(context.Background(), username); err != nil {
		return fmt.Errorf("EC2 인스턴스 시작 실패: %w", err)
	}

	hostname := username + "_VM"
	if err := s.repo.UpdateStatus(hostname, "running"); err != nil {
		return fmt.Errorf("DB 상태 업데이트 실패: %w", err)
	}

	return nil
}

func (s *HostingService) StopVM(email string) error {
	username := removeDomain(email)

	if err := s.EC2.StopVMByUser(context.Background(), username); err != nil {
		return fmt.Errorf("EC2 인스턴스 중지 실패: %w", err)
	}

	hostname := username + "_VM"
	if err := s.repo.UpdateStatus(hostname, "stopped"); err != nil {
		return fmt.Errorf("DB 상태 업데이트 실패: %w", err)
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
