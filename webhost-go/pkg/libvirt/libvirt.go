// Package libvirt provides functionality to manage VMs using libvirt directly in Go
package libvirt

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/digitalocean/go-libvirt"
	"net"
)

type LibvirtManager struct {
	conn *libvirt.Libvirt
}

type DomainInfo struct {
	State     uint8  // 현재 VM 상태 (e.g. Running, Shutoff 등)
	MaxMem    uint64 // 할당된 최대 메모리 (KB)
	Memory    uint64 // 현재 사용 중인 메모리 (KB)
	NrVirtCpu uint16 // 가상 CPU 수
	CpuTime   uint64 // 누적 CPU 사용 시간 (나노초)
}

func NewLibvirtManager() (*LibvirtManager, error) {
	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 2*time.Second)
	if err != nil {
		return nil, err
	}
	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		return nil, err
	}
	return &LibvirtManager{conn: l}, nil
}

func (m *LibvirtManager) CreateDisk(path string, sizeGB int) error {
	if _, err := os.Stat(path); err == nil {
		return nil // already exists
	}
	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", path, fmt.Sprintf("%dG", sizeGB))
	return cmd.Run()
}

func (m *LibvirtManager) LoadDomainXML(tmplPath string, cfg VMConfig) (string, error) {
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, cfg)
	return buf.String(), err
}

func (m *LibvirtManager) DefineAndStartVM(xmlStr string) error {
	dom, err := m.conn.DomainDefineXML(xmlStr)
	if err != nil {
		return err
	}
	return m.conn.DomainCreate(dom)
}

func (m *LibvirtManager) EnableSerialConsole(xmlTemplate string) string {
	// This would add a <console> tag, assuming template does not have it already
	// You should customize your XML template to include it instead if needed
	return strings.Replace(xmlTemplate, "</devices>", `
		<console type="pty">
		  <target type="serial" port="0"/>
		</console>
	</devices>`, 1)
}

func (m *LibvirtManager) Shutdown(name string) error {
	dom, err := m.conn.DomainLookupByName(name)
	if err != nil {
		return err
	}
	return m.conn.DomainShutdown(dom)
}

func (m *LibvirtManager) Destroy(name string) error {
	dom, err := m.conn.DomainLookupByName(name)
	if err != nil {
		return err
	}
	return m.conn.DomainDestroy(dom)
}

func (m *LibvirtManager) DeleteDomain(name string, withDisks bool) error {
	dom, err := m.conn.DomainLookupByName(name)
	if err != nil {
		return fmt.Errorf("도메인 조회 실패: %w", err)
	}

	// 실행 중이면 종료
	active, err := m.conn.DomainIsActive(dom)
	if err == nil && active != 0 {
		if err := m.conn.DomainDestroy(dom); err != nil {
			return fmt.Errorf("실행 중인 도메인 종료 실패: %w", err)
		}
	}

	var xmlDesc string
	if withDisks {
		xmlDesc, err = m.conn.DomainGetXMLDesc(dom, 0)
		if err != nil {
			return fmt.Errorf("도메인 XML 가져오기 실패: %w", err)
		}
	}

	// 정의 제거
	if err := m.conn.DomainUndefine(dom); err != nil {
		return fmt.Errorf("도메인 정의 삭제 실패: %w", err)
	}

	// 디스크 경로 수집 후 삭제
	if withDisks {
		paths := extractDiskPaths(xmlDesc)
		for _, path := range paths {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("디스크 삭제 실패 (%s): %w", path, err)
			}
		}
	}

	return nil
}

func (l *LibvirtManager) Resume(domainName string) error {
	domain, err := l.conn.DomainLookupByName(domainName)
	if err != nil {
		return fmt.Errorf("도메인 조회 실패: %w", err)
	}

	if err := l.conn.DomainResume(domain); err != nil {
		return fmt.Errorf("도메인 재개 실패: %w", err)
	}

	return nil
}

func (l *LibvirtManager) DomainIsActive(name string) (bool, error) {
	domain, err := l.conn.DomainLookupByName(name)
	if err != nil {
		return false, err
	}

	rState, _, _, _, _, _ := l.conn.DomainGetInfo(domain)
	if err != nil {
		return false, err
	}

	return rState == uint8(libvirt.DomainRunning), nil
}

func (l *LibvirtManager) GetDomainInfoByName(name string) (*DomainInfo, error) {
	// 도메인 조회
	domain, err := l.conn.DomainLookupByName(name)
	if err != nil {
		return nil, fmt.Errorf("도메인 조회 실패: %w", err)
	}

	// 도메인 정보 조회
	rState, rMaxMem, rMemory, rVCPU, rCPUTime, err := l.conn.DomainGetInfo(domain)
	if err != nil {
		return nil, fmt.Errorf("도메인 정보 조회 실패: %w", err)
	}

	info := &DomainInfo{
		State:     uint8(rState),
		MaxMem:    rMaxMem,
		Memory:    rMemory,
		NrVirtCpu: uint16(rVCPU),
		CpuTime:   rCPUTime,
	}

	return info, nil
}

type domainDiskXML struct {
	Disks []struct {
		Source struct {
			File string `xml:"file,attr"`
		} `xml:"source"`
	} `xml:"devices>disk"`
}

func extractDiskPaths(xmlDesc string) []string {
	var parsed domainDiskXML
	if err := xml.Unmarshal([]byte(xmlDesc), &parsed); err != nil {
		return nil
	}

	var paths []string
	for _, disk := range parsed.Disks {
		if disk.Source.File != "" {
			paths = append(paths, disk.Source.File)
		}
	}
	return paths
}

func (m *LibvirtManager) CopyTemplateDisk(dstPath string) error {
	srcPath := "/var/lib/libvirt/images/templates/ubuntu-22.04-server-cloudimg-amd64.img"
	cmd := exec.Command("cp", "-f", srcPath, dstPath)
	return cmd.Run()
}

func (m *LibvirtManager) StartUbuntuVM(vmName string) error {
	baseDir := filepath.Join("/var/lib/libvirt/images/instances", vmName)
	diskPath := filepath.Join(baseDir, "disk.qcow2")
	isoPath := filepath.Join(baseDir, "cloud-init.iso")
	xmlTemplatePath := "/etc/libvirt/templates/domain_template.xml"

	// 1. 디렉토리 생성
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return fmt.Errorf("VM 디렉토리 생성 실패: %w", err)
	}

	// 2. 루트 디스크 복사 (템플릿에서)
	if err := m.CopyTemplateDisk(diskPath); err != nil {
		return fmt.Errorf("디스크 복사 실패: %w", err)
	}

	// 3. cloud-init ISO 생성
	isoPath, err := m.CreateCloudInitISO(vmName)
	if err != nil {
		return fmt.Errorf("cloud-init ISO 생성 실패: %w", err)
	}

	// 4. 도메인 XML 로드 및 구성
	cfg := VMConfig{
		Name:     vmName,
		MemoryMB: 1024, // 메모리 설정 (1GB)
		VCPUs:    1,    // CPU 코어 수
		DiskPath: diskPath,
		ISOPath:  isoPath,
	}
	xmlStr, err := m.LoadDomainXML(xmlTemplatePath, cfg)
	if err != nil {
		return fmt.Errorf("도메인 XML 로드 실패: %w", err)
	}

	// 5. 시리얼 콘솔 추가
	xmlStr = m.EnableSerialConsole(xmlStr)

	// 6. 도메인 정의 및 시작
	if err := m.DefineAndStartVM(xmlStr); err != nil {
		return fmt.Errorf("VM 시작 실패: %w", err)
	}

	return nil
}

func (m *LibvirtManager) StartUbuntuVMWithStaticIP(vmName string, staticIP net.IP) error {
	baseDir := filepath.Join("/var/lib/libvirt/images/instances", vmName)
	diskPath := filepath.Join(baseDir, "disk.qcow2")
	isoPath := filepath.Join(baseDir, "cloud-init.iso")
	xmlTemplatePath := "/etc/libvirt/templates/domain_template.xml"

	// 1. 디렉토리 생성
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return fmt.Errorf("VM 디렉토리 생성 실패: %w", err)
	}

	// 2. 템플릿 디스크 복사
	if err := m.CopyTemplateDisk(diskPath); err != nil {
		return fmt.Errorf("디스크 복사 실패: %w", err)
	}

	// 디스크 크기 조절 (예: 10GB로 확장)
	resizeCmd := exec.Command("qemu-img", "resize", diskPath, "10G")
	if output, err := resizeCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("디스크 크기 조절 실패: %w\n출력: %s", err, output)
	}

	// 3. Static IP 기반 cloud-init ISO 생성
	isoPath, err := m.CreateCloudInitISOWithStaticIP(vmName, staticIP)
	if err != nil {
		return fmt.Errorf("cloud-init ISO 생성 실패: %w", err)
	}

	// 4. 도메인 XML 생성
	cfg := VMConfig{
		Name:     vmName,
		MemoryMB: 1024,
		VCPUs:    1,
		DiskPath: diskPath,
		ISOPath:  isoPath,
	}
	xmlStr, err := m.LoadDomainXML(xmlTemplatePath, cfg)
	if err != nil {
		return fmt.Errorf("도메인 XML 로드 실패: %w", err)
	}

	// 5. 시리얼 콘솔 추가
	xmlStr = m.EnableSerialConsole(xmlStr)

	// 6. VM 정의 및 시작
	if err := m.DefineAndStartVM(xmlStr); err != nil {
		return fmt.Errorf("VM 시작 실패: %w", err)
	}

	return nil
}
