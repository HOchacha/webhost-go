package token_test

import (
	"testing"
	"time"

	"webhost-go/webhost-go/internal/services/user_service/authn/token"
)

func TestJWTManager_ValidToken(t *testing.T) {
	manager := token.NewJWTManager("secret-key", time.Minute)
	email := "user@example.com"
	role := token.RoleAdmin

	tok, err := manager.Generate(email, role)
	if err != nil {
		t.Fatalf("토큰 생성 실패: %v", err)
	}

	result := manager.Validate(tok)
	if !result.Valid {
		t.Errorf("유효한 토큰이 검증에 실패함: %v", result.ParseErr)
	}
	if result.Claims.Email != email {
		t.Errorf("이메일이 일치하지 않음: got %s, want %s", result.Claims.Email, email)
	}
	if result.Claims.Role != role {
		t.Errorf("역할이 일치하지 않음: got %s, want %s", result.Claims.Role, role)
	}
}

func TestJWTManager_ExpiredToken(t *testing.T) {
	manager := token.NewJWTManager("secret-key", -time.Second) // 즉시 만료
	email := "expired@example.com"
	role := token.RoleUser

	tok, err := manager.Generate(email, role)
	if err != nil {
		t.Fatalf("토큰 생성 실패: %v", err)
	}

	result := manager.Validate(tok)
	if !result.Expired {
		t.Errorf("만료된 토큰인데 expired=false")
	}
	if result.Valid {
		t.Errorf("만료된 토큰이 valid=true로 판단됨")
	}
}

func TestJWTManager_InvalidSignature(t *testing.T) {
	manager := token.NewJWTManager("correct-secret", time.Minute)

	// 다른 키로 서명한 토큰
	otherManager := token.NewJWTManager("wrong-secret", time.Minute)
	tok, err := otherManager.Generate("user@example.com", "user")
	if err != nil {
		t.Fatalf("토큰 생성 실패: %v", err)
	}

	result := manager.Validate(tok)
	if result.Valid {
		t.Error("서명이 잘못된 토큰인데 valid=true")
	}
	if result.ParseErr == nil {
		t.Error("ParseErr가 nil인데 잘못된 서명임")
	}
}

func TestJWTManager_InvalidRole(t *testing.T) {
	manager := token.NewJWTManager("secret-key", time.Minute)
	email := "hacker@example.com"
	invalidRole := token.Role("superadmin") // 정의되지 않은 Role

	_, err := manager.Generate(email, invalidRole)
	if err == nil {
		t.Errorf("유효하지 않은 Role로 토큰이 생성됨: %s", invalidRole)
	}
}

func TestJWTManager_OnlyAdminCanAccess(t *testing.T) {
	manager := token.NewJWTManager("secret-key", time.Minute)

	tests := []struct {
		email       string
		role        token.Role
		shouldAllow bool
	}{
		{"admin@example.com", token.RoleAdmin, true},
		{"user@example.com", token.RoleUser, false},
		{"guest@example.com", token.RoleReader, false},
		{"hacker@example.com", token.Role("superadmin"), false},
	}

	for _, tc := range tests {
		tokenStr, err := manager.Generate(tc.email, tc.role)
		if err != nil {
			t.Errorf("토큰 생성 실패 (%s): %v", tc.role, err)
			continue
		}

		result := manager.Validate(tokenStr)
		if !result.Valid {
			t.Errorf("토큰 검증 실패 (%s): %v", tc.role, result.ParseErr)
			continue
		}

		if result.Claims.Role == token.RoleAdmin && !tc.shouldAllow {
			t.Errorf("admin은 아닌데 통과함: %s", result.Claims.Role)
		}
		if result.Claims.Role != token.RoleAdmin && tc.shouldAllow {
			t.Errorf("admin인데 거부됨: %s", result.Claims.Role)
		}
	}
}
