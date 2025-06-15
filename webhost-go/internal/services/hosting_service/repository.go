package hosting_service

type HostingRepository interface {
	Create(h *Hosting) error
	UpdateStatus(vmName string, status string) error
	Delete(vmName string) error
	FindByVMName(vmName string) (*Hosting, error)
	FindAllByUserID(userID int64) ([]*Hosting, error)
	FindAll() ([]*Hosting, error) // ✅ 모든 VM 조회 추가
	GetAvailablePort(basePort, maxPort int) (int, error)
	FindActiveByUserID(userID int64) (*Hosting, error)
	GetUsedIPs() ([]string, error)
}
