package token_test

import (
	"testing"
	"time"

	"webhost-go/webhost-go/internal/services/user_service/authn/token"
)

func TestJWTManager_ValidToken(t *testing.T) {
	manager := token.NewJWTManager("secret-key", time.Minute)
	email := "user@example.com"
	role := "admin"

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
	role := "user"

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
