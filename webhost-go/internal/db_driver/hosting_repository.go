package db_driver

import (
	"database/sql"
	"errors"
	"fmt"
	"webhost-go/webhost-go/internal/services/hosting_service"
)

type HostingRepository struct {
	db *sql.DB
}

func NewHostingRepository(db *sql.DB) *HostingRepository {
	return &HostingRepository{db: db}
}

func (r *HostingRepository) Create(h *hosting_service.Hosting) error {
	_, err := r.db.Exec(`
		INSERT INTO hostings (user_id, vm_name, ip_address, ssh_port, proxy_path, disk_path, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, h.UserID, h.VMName, h.IPAddress, h.SSHPort, h.ProxyPath, h.DiskPath, h.Status)
	return err
}

func (r *HostingRepository) UpdateStatus(vmName string, status string) error {
	_, err := r.db.Exec(`
		UPDATE hostings SET status = ? WHERE vm_name = ?
	`, status, vmName)
	return err
}

func (r *HostingRepository) Delete(vmName string) error {
	_, err := r.db.Exec(`
		DELETE FROM hostings WHERE vm_name = ?
	`, vmName)
	return err
}

func (r *HostingRepository) FindByVMName(vmName string) (*hosting_service.Hosting, error) {
	row := r.db.QueryRow(`
		SELECT id, user_id, vm_name, ip_address, ssh_port, proxy_path, disk_path, status, created_at
		FROM hostings
		WHERE vm_name = ? AND status != 'deleted'
	`, vmName)

	var h hosting_service.Hosting
	if err := row.Scan(
		&h.ID, &h.UserID, &h.VMName, &h.IPAddress,
		&h.SSHPort, &h.ProxyPath, &h.DiskPath,
		&h.Status, &h.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &h, nil
}

func (r *HostingRepository) FindAllByUserID(userID int64) ([]*hosting_service.Hosting, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, vm_name, ip_address, ssh_port, proxy_path, disk_path, status, created_at
		FROM hostings WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*hosting_service.Hosting
	for rows.Next() {
		var h hosting_service.Hosting
		if err := rows.Scan(
			&h.ID, &h.UserID, &h.VMName, &h.IPAddress,
			&h.SSHPort, &h.ProxyPath, &h.DiskPath,
			&h.Status, &h.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &h)
	}
	return list, nil
}

func (r *HostingRepository) FindAll() ([]*hosting_service.Hosting, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, vm_name, ip_address, ssh_port, proxy_path, disk_path, status, created_at
		FROM hostings
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hostings []*hosting_service.Hosting
	for rows.Next() {
		var h hosting_service.Hosting
		if err := rows.Scan(
			&h.ID, &h.UserID, &h.VMName, &h.IPAddress,
			&h.SSHPort, &h.ProxyPath, &h.DiskPath,
			&h.Status, &h.CreatedAt,
		); err != nil {
			return nil, err
		}
		hostings = append(hostings, &h)
	}
	return hostings, nil
}

func (r *HostingRepository) GetAvailablePort(basePort, maxPort int) (int, error) {
	rows, err := r.db.Query("SELECT ssh_port FROM hostings WHERE status != 'deleted'")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	used := make(map[int]bool)
	for rows.Next() {
		var port int
		if err := rows.Scan(&port); err != nil {
			return 0, err
		}
		used[port] = true
	}

	for p := basePort; p <= maxPort; p++ {
		if !used[p] {
			return p, nil
		}
	}

	return 0, fmt.Errorf("사용 가능한 포트를 찾을 수 없습니다")
}

func (r *HostingRepository) FindActiveByUserID(userID int64) (*hosting_service.Hosting, error) {
	row := r.db.QueryRow(`
		SELECT id, user_id, vm_name, ip_address, ssh_port, proxy_path, disk_path, status, created_at
		FROM hostings
		WHERE user_id = ? AND status != 'deleted'
	`, userID)

	var h hosting_service.Hosting
	if err := row.Scan(
		&h.ID, &h.UserID, &h.VMName, &h.IPAddress,
		&h.SSHPort, &h.ProxyPath, &h.DiskPath,
		&h.Status, &h.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &h, nil
}

func (r *HostingRepository) GetUsedIPs() ([]string, error) {
	rows, err := r.db.Query(`
		SELECT ip_address FROM hostings WHERE status != 'deleted'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ips []string
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err != nil {
			return nil, err
		}
		ips = append(ips, ip)
	}
	return ips, nil
}
