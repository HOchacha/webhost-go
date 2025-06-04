package utils_test

import (
	"log"
	"testing"

	"webhost-go/webhost-go/pkg/libvirt/utils"
)

func TestCreateAndStartVM(t *testing.T) {
	diskPath := "/var/lib/libvirt/disks/test-vm.qcow2"
	if err := utils.CreateDiskIfNotExists(diskPath, 5); err != nil {
		t.Fatalf("디스크 생성 실패: %v", err)
	}

	// 1. 템플릿 파라미터 정의
	params := utils.DomainParams{
		Name:     "test-vm",
		Memory:   512,
		VCPU:     1,
		DiskPath: diskPath,
	}

	// 2. XML 템플릿 생성
	xml, err := utils.LoadDomainXML("domain_template.xml", params)
	if err != nil {
		t.Fatalf("fail to load Template: %v", err)
	}

	// 3. libvirt 연결
	l, err := utils.ConnectLibvirt()
	if err != nil {
		t.Fatalf("Fail to connect to libvirt: %v", err)
	}
	defer l.Disconnect()

	if err := utils.UndefineVMIfExists(l, params.Name); err != nil {
		t.Fatalf("fail to remove existing test domain: %v", err)
	}

	// 4. VM 생성 및 실행
	if err := utils.CreateAndStartVM(l, xml); err != nil {
		t.Fatalf("Fail to VM Create and Start: %v", err)
	}

	log.Println("✅  Test Success: VM Creation and Execution")

	// 5. 테스트 후 정리
	if err := utils.ShutdownVMByName(l, params.Name); err != nil {
		t.Logf("⚠️  Warning: graceful shutdown failed, trying force destroy: %v", err)
		_ = utils.DestroyVMByName(l, params.Name)
	}

	// (optional) 완전 제거하고 싶을 경우
	if err := utils.UndefineVMIfExists(l, params.Name); err != nil {
		t.Logf("⚠️  Warning: undefine failed: %v", err)
	}

	log.Println("✅  Test Success: VM destroyed")
}
