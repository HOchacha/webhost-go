package hosting_service

import "time"

type Hosting struct {
	ID        int64  // 내부 DB용 ID
	UserID    int64  // 소유자
	VMName    string // libvirt 도메인 이름
	IPAddress string // VM의 내부 IP 주소
	SSHPort   int    // 외부에서 접속 가능한 SSH 포트 (nginx stream용)
	ProxyPath string
	DiskPath  string // qcow2 디스크 경로
	Status    string // Running, Stopped, Error 등
	CreatedAt time.Time
}

type HostingPlan struct {
	Name     string // "small", "medium", "large"
	CPU      int
	MemoryMB int
	DiskGB   int
}

var DefaultPlans = map[string]HostingPlan{
	"small":  {Name: "small", CPU: 1, MemoryMB: 1024, DiskGB: 10},
	"medium": {Name: "medium", CPU: 2, MemoryMB: 2048, DiskGB: 20},
	"large":  {Name: "large", CPU: 4, MemoryMB: 4096, DiskGB: 40},
}
