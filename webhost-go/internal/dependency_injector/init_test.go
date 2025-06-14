package dependency_injector_test

import (
	_ "github.com/go-sql-driver/mysql"
	"testing"
	"time"
	"webhost-go/webhost-go/internal/dependency_injector"
)

func TestInitApp(t *testing.T) {
	ai := &dependency_injector.AppInitializer{
		DB: dependency_injector.DBConfig{
			User:     "testuser",
			Password: "testpass",
			Host:     "127.0.0.1",
			Port:     3306,
			Name:     "testdb",
		},
		JWTSecret: "test-secret",
		TokenTTL:  time.Minute * 15,
	}

	reg, err := ai.InitApp()
	if err != nil {
		t.Fatalf("InitApp failed: %v", err)
	}

	if reg.UserHandler == nil {
		t.Error("UserHandler is nil")
	}
	if reg.JWTManager == nil {
		t.Error("JWTManager is nil")
	}
	if reg.AuthMiddleware == nil {
		t.Error("AuthMiddleware is nil")
	}
}
