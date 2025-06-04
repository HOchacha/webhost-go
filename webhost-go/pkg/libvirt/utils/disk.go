package utils

import (
	"fmt"
	"os"
	"os/exec"
)

// CreateDiskIfNotExists 생성할 디스크 경로와 크기(GB)를 받아 qcow2 디스크를 생성
func CreateDiskIfNotExists(path string, sizeGB int) error {
	if _, err := os.Stat(path); err == nil {
		// 디스크가 이미 존재함
		return nil
	}

	cmd := exec.Command("qemu-img", "create", "-f", "qcow2", path, fmt.Sprintf("%dG", sizeGB))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create disk image: %w", err)
	}

	return nil
}
