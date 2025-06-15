package libvirt_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"webhost-go/webhost-go/pkg/libvirt"
)

func TestLibvirtVMCreation(t *testing.T) {
	manager, err := libvirt.NewLibvirtManager()
	if err != nil {
		t.Fatalf("failed to connect to libvirt: %v", err)
	}

	vmName := "testvm-" + time.Now().Format("150405")
	diskPath := filepath.Join("/var/lib/libvirt/images/", vmName+".qcow2")
	isoPath := filepath.Join("/var/lib/libvirt/images/cloud-init", vmName+".iso")
	xmlTemplate := "/etc/libvirt/templates/domain_template.xml" // 실제 경로 맞게 수정

	// Clean up after test
	t.Cleanup(func() {
		_ = manager.Destroy(vmName)
		_ = os.Remove(diskPath)
		_ = os.Remove(isoPath)
	})

	// Step 1: Create Disk
	err = manager.CreateDisk(diskPath, 5)
	if err != nil {
		t.Fatalf("failed to create disk: %v", err)
	}

	// Step 2: Create cloud-init ISO
	isoPath, err = manager.CreateCloudInitISO(vmName)
	if err != nil {
		t.Fatalf("failed to create cloud-init ISO: %v", err)
	}

	// Step 3: Load and render XML
	cfg := libvirt.VMConfig{
		Name:     vmName,
		MemoryMB: 2028,
		VCPUs:    2,
		DiskPath: diskPath,
		ISOPath:  isoPath,
	}

	xmlStr, err := manager.LoadDomainXML(xmlTemplate, cfg)
	if err != nil {
		t.Fatalf("failed to load domain XML: %v", err)
	}
	xmlStr = manager.EnableSerialConsole(xmlStr)

	// Step 4: Define and start VM
	err = manager.DefineAndStartVM(xmlStr)
	if err != nil {
		t.Fatalf("failed to start VM: %v", err)
	}
}

func TestStartUbuntuVM(t *testing.T) {
	manager, err := libvirt.NewLibvirtManager()
	if err != nil {
		t.Fatalf("libvirt 연결 실패: %v", err)
	}

	// VM 이름을 현재 시간 기반으로 설정
	vmName := "testvm-" + time.Now().Format("150405")

	// 테스트 후 정리
	t.Cleanup(func() {
		err := manager.DeleteDomain(vmName, true)
		if err != nil {
			t.Logf("VM 삭제 중 오류 발생: %v", err)
		}

		// 디렉토리 정리
		vmDir := filepath.Join("/var/lib/libvirt/images/instances", vmName)
		_ = os.RemoveAll(vmDir)
	})

	// VM 생성 및 시작 테스트
	if err := manager.StartUbuntuVM(vmName); err != nil {
		t.Fatalf("Ubuntu VM 생성 실패: %v", err)
	}
}

func TestStartUbuntuVMWithStaticIP(t *testing.T) {
	manager, err := libvirt.NewLibvirtManager()
	if err != nil {
		t.Fatalf("libvirt 연결 실패: %v", err)
	}

	// 사용 가능한 IP 목록 조회
	ips, err := manager.GetUsableIPs()
	if err != nil || len(ips) == 0 {
		t.Fatalf("사용 가능한 IP 조회 실패 또는 없음: %v", err)
	}

	vmName := "testvm-static-" + time.Now().Format("150405")
	ip := ips[0]

	err = manager.StartUbuntuVMWithStaticIP(vmName, ip)
	if err != nil {
		t.Fatalf("Static IP 기반 VM 생성 실패: %v", err)
	}

	// 확인: 디스크와 ISO 파일 존재하는지 확인
	baseDir := filepath.Join("/var/lib/libvirt/images/instances", vmName)
	if _, err := os.Stat(filepath.Join(baseDir, "disk.qcow2")); err != nil {
		t.Errorf("디스크 파일 없음: %v", err)
	}
	isoPath := filepath.Join(baseDir, vmName+".iso")
	if _, err := os.Stat(isoPath); err != nil {
		t.Errorf("ISO 파일 없음: %v", err)
	}

	// 테스트 종료 후 정리
	t.Cleanup(func() {
		_ = manager.DeleteDomain(vmName, true)
		_ = os.RemoveAll(baseDir)
	})
}
