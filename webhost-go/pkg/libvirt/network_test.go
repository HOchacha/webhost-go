package libvirt_test

import (
	"fmt"
	"testing"

	"webhost-go/webhost-go/pkg/libvirt"
)

func TestGetDefaultNetworkConfig(t *testing.T) {
	manager, err := libvirt.NewLibvirtManager()
	if err != nil {
		t.Fatalf("libvirt 연결 실패: %v", err)
	}

	conf, err := manager.GetDefaultNetworkConfig()
	if err != nil {
		t.Fatalf("기본 네트워크 설정 조회 실패: %v", err)
	}

	expectedCIDR := "192.168.122.0/24"
	expectedGateway := "192.168.122.1"

	if conf.CIDR.String() != expectedCIDR {
		t.Errorf("CIDR mismatch: got %s, want %s", conf.CIDR.String(), expectedCIDR)
	}
	if conf.Gateway.String() != expectedGateway {
		t.Errorf("Gateway mismatch: got %s, want %s", conf.Gateway.String(), expectedGateway)
	}
}

func TestGetUsableIPs(t *testing.T) {
	manager, err := libvirt.NewLibvirtManager()
	if err != nil {
		t.Fatalf("libvirt 연결 실패: %v", err)
	}

	// CIDR 및 Gateway 조회
	netConf, err := manager.GetDefaultNetworkConfig()
	if err != nil {
		t.Fatalf("기본 네트워크 구성 조회 실패: %v", err)
	}

	fmt.Println("📦 CIDR:", netConf.CIDR.String())
	fmt.Println("🌐 Gateway:", netConf.Gateway.String())

	// 사용 가능한 IP 목록 가져오기
	usableIPs, err := manager.GetUsableIPs()
	if err != nil {
		t.Fatalf("사용 가능한 IP 조회 실패: %v", err)
	}

	fmt.Printf("✅ 사용 가능한 IP (%d개):\n", len(usableIPs))
	for i, ip := range usableIPs {
		if i >= 10 {
			fmt.Println("... 생략")
			break
		}
		fmt.Println("-", ip)
	}
}
