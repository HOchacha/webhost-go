package agent

import "webhost-go/webhost-go/pkg/libvirt/utils"

type VMRequest struct {
	Name     string
	MemoryMB int
	VCPU     int
	DiskPath string
}

func StartVM(req VMRequest) error {
	// 1. 디스크 생성
	if err := utils.CreateDiskIfNotExists(req.DiskPath, 5); err != nil {
		return err
	}

	// 2. libvirt 연결
	conn, err := utils.ConnectLibvirt()
	if err != nil {
		return err
	}
	defer conn.Disconnect()

	// 3. 기존 VM 제거
	if err := utils.UndefineVMIfExists(conn, req.Name); err != nil {
		return err
	}

	// 4. XML 생성
	xmlParams := utils.DomainParams{
		Name:     req.Name,
		Memory:   req.MemoryMB,
		VCPU:     req.VCPU,
		DiskPath: req.DiskPath,
	}
	xmlStr, err := utils.LoadDomainXML("vm_agent/template.xml", xmlParams)
	if err != nil {
		return err
	}

	// 5. VM 생성 및 시작
	return utils.CreateAndStartVM(conn, xmlStr)
}

func StopVM(name string) error {
	conn, err := utils.ConnectLibvirt()
	if err != nil {
		return err
	}
	defer conn.Disconnect()

	return utils.ShutdownVMByName(conn, name)
}

func DeleteVM(name string) error {
	conn, err := utils.ConnectLibvirt()
	if err != nil {
		return err
	}
	defer conn.Disconnect()

	return utils.UndefineVMIfExists(conn, name)
}
