package hosting_service

import (
	"webhost-go/webhost-go/pkg/libvirt"
)

type Service interface {
	CreateHosting(userID int64, email string) (*Hosting, error)
	DeleteVM(name string) error
	GetVMStatus(name string) (*VMStatus, error)
	GetVMDetail(name string) (*Hosting, *libvirt.DomainInfo, error)
	StartVM(name string) error
	StopVM(name string) error
}
