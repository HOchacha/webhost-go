package cloudinit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func GenerateCloudInitISO(vmID, userID, password, hostname string) error {
	basePath := filepath.Join("/var/lib/libvirt/cloud-init", vmID)
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return fmt.Errorf("cannot create base directory: %w", err)
	}

	userData := fmt.Sprintf(`#cloud-config
users:
  - name: %s
    lock_passwd: false
    plain_text_passwd: '%s'
    shell: /bin/bash
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
chpasswd:
  expire: false
ssh_pwauth: true
`, userID, password)

	metaData := fmt.Sprintf(`instance-id: %s
local-hostname: %s
`, vmID, hostname)

	if err := os.WriteFile(filepath.Join(basePath, "user-data"), []byte(userData), 0644); err != nil {
		return fmt.Errorf("failed to write user-data: %w", err)
	}
	if err := os.WriteFile(filepath.Join(basePath, "meta-data"), []byte(metaData), 0644); err != nil {
		return fmt.Errorf("failed to write meta-data: %w", err)
	}

	isoPath := filepath.Join(basePath, "seed.iso")
	cmd := exec.Command("genisoimage", "-output", isoPath, "-volid", "cidata", "-joliet", "-rock", "user-data", "meta-data")
	cmd.Dir = basePath

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create ISO: %v\nOutput: %s", err, out)
	}

	return nil
}
