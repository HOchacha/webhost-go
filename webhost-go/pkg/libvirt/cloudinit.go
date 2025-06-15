package libvirt

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
)

func (m *LibvirtManager) CreateCloudInitISO(vmName string) (string, error) {
	baseDir := filepath.Join("/var/lib/libvirt/images/instances", vmName)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", err
	}

	metaPath := filepath.Join(baseDir, "meta-data")
	userDataPath := "/var/lib/libvirt/images/cloud-init/share/user-data" // 공통
	isoPath := filepath.Join(baseDir, "cloud-init.iso")

	meta := fmt.Sprintf("instance-id: %s\nlocal-hostname: %s\n", vmName, vmName)
	if err := os.WriteFile(metaPath, []byte(meta), 0644); err != nil {
		return "", err
	}

	cmd := exec.Command("genisoimage", "-output", isoPath, "-volid", "cidata", "-joliet", "-rock", userDataPath, metaPath)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return isoPath, nil
}

func (m *LibvirtManager) CreateCloudInitISOWithStaticIP(vmName string, ip net.IP) (string, error) {
	baseDir := filepath.Join("/var/lib/libvirt/images/instances", vmName)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", fmt.Errorf("디렉터리 생성 실패: %w", err)
	}

	// 기본 네트워크 설정에서 gateway, CIDR 가져오기
	netConf, err := m.GetDefaultNetworkConfig()
	if err != nil {
		return "", fmt.Errorf("네트워크 설정 조회 실패: %w", err)
	}
	maskSize, _ := netConf.CIDR.Mask.Size()

	// 1. meta-data 생성
	metaPath := filepath.Join(baseDir, "meta-data")
	meta := fmt.Sprintf("instance-id: %s\nlocal-hostname: %s\n", vmName, vmName)
	if err := os.WriteFile(metaPath, []byte(meta), 0644); err != nil {
		return "", fmt.Errorf("meta-data 작성 실패: %w", err)
	}

	// 2. network-config 생성
	networkPath := filepath.Join(baseDir, "network-config")
	networkYaml := fmt.Sprintf(`version: 2
ethernets:
  enp0s2:
    dhcp4: false
    addresses: [%s/%d]
    gateway4: %s
    nameservers:
      addresses: [8.8.8.8, 1.1.1.1]
`, ip.String(), maskSize, netConf.Gateway.String())
	if err := os.WriteFile(networkPath, []byte(networkYaml), 0644); err != nil {
		return "", fmt.Errorf("network-config 작성 실패: %w", err)
	}

	// 3. user-data는 공유 경로에서
	userDataPath := "/var/lib/libvirt/images/cloud-init/share/user-data"
	if _, err := os.Stat(userDataPath); err != nil {
		return "", fmt.Errorf("user-data 경로 확인 실패: %w", err)
	}

	// 4. ISO 생성
	isoPath := filepath.Join(baseDir, vmName+".iso")
	cmd := exec.Command("genisoimage",
		"-output", isoPath,
		"-volid", "cidata",
		"-joliet", "-rock",
		userDataPath, metaPath, networkPath,
	)

	var output []byte
	output, err = cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ISO 생성 실패: %w\n명령 출력: %s", err, string(output))
	}
	fmt.Println(string(output))
	return isoPath, nil
}
