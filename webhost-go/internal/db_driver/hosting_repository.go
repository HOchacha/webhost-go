package db_driver

import (
	"database/sql"
	"errors"
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
		INSERT INTO hostings 
		(user_id, name, plan, status, ip_address, ssh_port, disk_path, node_name, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, h.UserID, h.Name, h.Plan, h.Status, h.IPAddress, h.SSHPort, h.DiskPath, h.NodeName, h.CreatedAt)
	return err
}

func (r *HostingRepository) Update(h *hosting_service.Hosting) error {
	_, err := r.db.Exec(`
		UPDATE hostings SET
			user_id = ?, name = ?, plan = ?, status = ?, ip_address = ?, 
			ssh_port = ?, disk_path = ?, node_name = ?
		WHERE id = ?
	`, h.UserID, h.Name, h.Plan, h.Status, h.IPAddress, h.SSHPort, h.DiskPath, h.NodeName, h.ID)
	return err
}

func (r *HostingRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM hostings WHERE id = ?", id)
	return err
}

func (r *HostingRepository) FindByID(id int64) (*hosting_service.Hosting, error) {
	row := r.db.QueryRow("SELECT id, user_id, name, plan, status, ip_address, ssh_port, disk_path, node_name, created_at FROM hostings WHERE id = ?", id)

	var h hosting_service.Hosting
	if err := row.Scan(
		&h.ID, &h.UserID, &h.Name, &h.Plan, &h.Status,
		&h.IPAddress, &h.SSHPort, &h.DiskPath, &h.NodeName, &h.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("hosting not found")
		}
		return nil, err
	}
	return &h, nil
}

func (r *HostingRepository) FindByUserID(userID int64) ([]*hosting_service.Hosting, error) {
	rows, err := r.db.Query("SELECT id, user_id, name, plan, status, ip_address, ssh_port, disk_path, node_name, created_at FROM hostings WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*hosting_service.Hosting
	for rows.Next() {
		var h hosting_service.Hosting
		if err := rows.Scan(
			&h.ID, &h.UserID, &h.Name, &h.Plan, &h.Status,
			&h.IPAddress, &h.SSHPort, &h.DiskPath, &h.NodeName, &h.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, &h)
	}
	return result, nil
}

func (r *HostingRepository) FindAll() ([]*hosting_service.Hosting, error) {
	rows, err := r.db.Query("SELECT id, user_id, name, plan, status, ip_address, ssh_port, disk_path, node_name, created_at FROM hostings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*hosting_service.Hosting
	for rows.Next() {
		var h hosting_service.Hosting
		if err := rows.Scan(
			&h.ID, &h.UserID, &h.Name, &h.Plan, &h.Status,
			&h.IPAddress, &h.SSHPort, &h.DiskPath, &h.NodeName, &h.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, &h)
	}
	return result, nil
}
