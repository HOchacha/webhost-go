package db_driver_test

import (
	"database/sql"
	"testing"
	"webhost-go/webhost-go/internal/db_driver"
	"webhost-go/webhost-go/internal/services/user_service"

	_ "github.com/go-sql-driver/mysql"
)

func setupTestDB(t *testing.T) *sql.DB {
	// 실제 테스트용 MariaDB DB 사용
	dsn := "testuser:testpass@tcp(127.0.0.1:3306)/testdb?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("DB 연결 실패: %v", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		email VARCHAR(255) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL,
		role VARCHAR(50) NOT NULL DEFAULT 'user',
		name VARCHAR(255) NOT NULL
	);`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("스키마 생성 실패: %v", err)
	}
	return db
}

func TestUserRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := db_driver.NewUserRepository(db)

	// Create
	u := &user_service.User{
		Email:    "test@example.com",
		Password: "hashed_password",
		Role:     "user",
		Name:     "Test User",
	}
	if err := repo.Create(u); err != nil {
		t.Fatalf("사용자 생성 실패: %v", err)
	}

	// FindByEmail
	found, err := repo.FindByEmail("test@example.com")
	if err != nil {
		t.Fatalf("FindByEmail 실패: %v", err)
	}
	if found.Email != u.Email {
		t.Errorf("FindByEmail 반환값 불일치")
	}

	// FindByID
	foundByID, err := repo.FindByID(found.ID)
	if err != nil {
		t.Fatalf("FindByID 실패: %v", err)
	}
	if foundByID.Name != u.Name {
		t.Errorf("FindByID 반환값 불일치")
	}

	// Update
	foundByID.Name = "Updated User"
	if err := repo.Update(foundByID); err != nil {
		t.Fatalf("Update 실패: %v", err)
	}
	updated, _ := repo.FindByID(foundByID.ID)
	if updated.Name != "Updated User" {
		t.Errorf("이름 업데이트 실패")
	}

	// FindAll
	all, err := repo.FindAll()
	if err != nil || len(all) != 1 {
		t.Fatalf("FindAll 실패: %v", err)
	}

	// Delete
	if err := repo.Delete(foundByID.ID); err != nil {
		t.Fatalf("Delete 실패: %v", err)
	}
	_, err = repo.FindByID(foundByID.ID)
	if err == nil {
		t.Errorf("삭제된 사용자 조회됨")
	}
}
