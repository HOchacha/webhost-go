package db_driver_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"webhost-go/webhost-go/internal/db_driver"
	"webhost-go/webhost-go/internal/services/hosting_service"
)

func setupTestDB_hosting(t *testing.T) *sql.DB {
	dsn := getTestDSN()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("failed to connect to DB: %v", err)
	}

	_, err = db.Exec(`DROP TABLE IF EXISTS hostings`)
	if err != nil {
		t.Fatalf("failed to clean up table: %v", err)
	}

	schema := `
	CREATE TABLE hostings (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id BIGINT NOT NULL,
		name VARCHAR(255) NOT NULL,
		plan VARCHAR(50) NOT NULL,
		status VARCHAR(50) NOT NULL,
		ip_address VARCHAR(255),
		ssh_port INT,
		disk_path VARCHAR(255),
		node_name VARCHAR(100),
		created_at DATETIME NOT NULL
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	return db
}

// 환경변수 또는 기본값을 통해 DSN 구성
func getTestDSN() string {
	user := getenv("TEST_DB_USER", "testuser")
	pass := getenv("TEST_DB_PASS", "testpass")
	host := getenv("TEST_DB_HOST", "127.0.0.1")
	port := getenv("TEST_DB_PORT", "3306")
	name := getenv("TEST_DB_NAME", "testdb")

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, name)
}

func getenv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func TestHostingRepository_MySQL(t *testing.T) {
	db := setupTestDB_hosting(t)
	repo := db_driver.NewHostingRepository(db)

	now := time.Now().UTC()
	h := &hosting_service.Hosting{
		UserID:    1,
		Name:      "mysql-vm",
		Plan:      "small",
		Status:    "Running",
		IPAddress: "192.168.0.123",
		SSHPort:   22222,
		DiskPath:  "/var/lib/libvirt/images/mysql-vm.qcow2",
		NodeName:  "node-a",
		CreatedAt: now,
	}

	// Create
	if err := repo.Create(h); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// FindAll
	all, err := repo.FindAll()
	if err != nil || len(all) != 1 {
		t.Fatalf("FindAll failed: %v", err)
	}

	// FindByID
	found, err := repo.FindByID(all[0].ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Name != "mysql-vm" {
		t.Errorf("expected name 'mysql-vm', got %s", found.Name)
	}

	// Update
	found.Status = "Stopped"
	if err := repo.Update(found); err != nil {
		t.Fatalf("update failed: %v", err)
	}
	updated, _ := repo.FindByID(found.ID)
	if updated.Status != "Stopped" {
		t.Errorf("update not applied, got status: %s", updated.Status)
	}

	// Delete
	if err := repo.Delete(found.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := repo.FindByID(found.ID); err == nil {
		t.Errorf("expected error after deletion, got nil")
	}
}
