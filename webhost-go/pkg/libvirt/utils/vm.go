package utils

import (
	"fmt"
	"github.com/digitalocean/go-libvirt"
	"log"
	"net"
	"time"
)

func ConnectLibvirt() (*libvirt.Libvirt, error) {
	conn, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 2*time.Second)
	if err != nil {
		return nil, err
	}
	l := libvirt.New(conn)
	if err := l.Connect(); err != nil {
		return nil, err
	}
	return l, nil
}

func CreateAndStartVM(l *libvirt.Libvirt, xml string) error {
	domain, err := l.DomainDefineXML(xml)
	if err != nil {
		return fmt.Errorf("define failed: %w", err)
	}
	if err := l.DomainCreate(domain); err != nil {
		return fmt.Errorf("start failed: %w", err)
	}
	return nil
}

func UndefineVMIfExists(l *libvirt.Libvirt, name string) error {
	dom, err := l.DomainLookupByName(name)
	if err != nil {
		// 도메인 없음 → 무시
		return nil
	}

	// Persistent 여부 확인 → destroy 전에 미리 확인!
	persistent, err := l.DomainIsPersistent(dom)
	if err != nil {
		return fmt.Errorf("failed to check persistent state: %w", err)
	}

	// 실행 중이면 종료 시도
	active, err := l.DomainIsActive(dom)
	if err == nil && active != 0 {
		_ = l.DomainDestroy(dom)
	}

	if persistent != 0 {
		if err := l.DomainUndefine(dom); err != nil {
			return fmt.Errorf("failed to undefine domain: %w", err)
		}
	} else {
		log.Printf("ℹ️  Domain '%s' is transient; skipping undefine", name)
	}

	return nil
}

func ShutdownVMByName(l *libvirt.Libvirt, name string) error {
	dom, err := l.DomainLookupByName(name)
	if err != nil {
		return fmt.Errorf("failed to lookup domain: %w", err)
	}
	if err := l.DomainShutdown(dom); err != nil {
		return fmt.Errorf("failed to shutdown domain: %w", err)
	}
	return nil
}

func DestroyVMByName(l *libvirt.Libvirt, name string) error {
	dom, err := l.DomainLookupByName(name)
	if err != nil {
		return fmt.Errorf("failed to lookup domain: %w", err)
	}
	if err := l.DomainDestroy(dom); err != nil {
		return fmt.Errorf("failed to destroy domain: %w", err)
	}
	return nil
}
