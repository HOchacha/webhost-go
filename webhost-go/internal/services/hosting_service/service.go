package hosting_service

type Service interface {
	CreateHosting(userID int64, email string) (*Hosting, error)
	DeleteVM(email string) error
	GetVMStatus(email string) (*VMStatus, error)
	GetVMDetail(email string) (*Hosting, *EC2InstanceInfo, error)
	StartVM(email string) error
	StopVM(email string) error
}
