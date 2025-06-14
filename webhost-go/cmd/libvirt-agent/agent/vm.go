package agent

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"webhost-go/webhost-go/pkg/libvirt/utils"
)

type VMRequest struct {
	Name     string
	MemoryMB int
	VCPU     int
	DiskPath string
}

// StartVM creates a disk from a template and launches a new VM
func StartVM(req VMRequest) error {
	// 1. 경로 설정
	diskPath := filepath.Join("/var/lib/libvirt/images", fmt.Sprintf("%s.qcow2", req.Name))
	templatePath := filepath.Join(TemplateDir, "base.qcow2")

	// 2. 기존 디스크가 없으면 템플릿 복사
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		if err := copyFile(templatePath, diskPath); err != nil {
			return fmt.Errorf("디스크 생성 실패: %w", err)
		}
	}

	// 3. libvirt 연결
	conn, err := utils.ConnectLibvirt()
	if err != nil {
		return fmt.Errorf("libvirt 연결 실패: %w", err)
	}
	defer conn.Disconnect()

	// 4. 기존 VM 제거 (중복 방지)
	if err := utils.UndefineVMIfExists(conn, req.Name); err != nil {
		return fmt.Errorf("기존 VM 제거 실패: %w", err)
	}

	// 5. XML 생성
	xmlParams := utils.DomainParams{
		Name:     req.Name,
		Memory:   req.MemoryMB,
		VCPU:     req.VCPU,
		DiskPath: diskPath,
	}
	xmlStr, err := utils.LoadDomainXML("/etc/libvirt-agent/template.xml", xmlParams)
	if err != nil {
		return fmt.Errorf("XML 생성 실패: %w", err)
	}

	// 6. VM 생성 및 시작
	if err := utils.CreateAndStartVM(conn, xmlStr); err != nil {
		return fmt.Errorf("VM 시작 실패: %w", err)
	}

	return nil
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

func CreateDiskFromTemplate(templatePath, targetPath string) error {
	input, err := os.Open(templatePath)
	if err != nil {
		return err
	}
	defer input.Close()

	out, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, input)
	return err
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
