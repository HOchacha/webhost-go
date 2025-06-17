package aws_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	myaws "webhost-go/webhost-go/pkg/aws" // ← 실제 경로로 수정
)

func TestEC2InstanceLifecycle(t *testing.T) {
	ctx := context.Background()

	// AWS 설정 (환경변수 또는 IAM 역할 기반)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Fatalf("AWS 설정 로딩 실패: %v", err)
	}

	manager := myaws.NewEC2Manager(cfg)

	// 1. 인스턴스 생성
	name := "test-ec2-" + time.Now().Format("150405")
	user := "testuser"

	t.Logf("1️⃣ 인스턴스 생성 시도: %s", name)
	instanceID, publicIP, _, err := manager.StartUbuntuVMWithOwner(ctx, name, user)
	if err != nil {
		t.Fatalf("인스턴스 생성 실패: %v", err)
	}
	t.Logf("✅ 생성 완료: ID=%s, IP=%s", instanceID, publicIP)

	// 👇 여기에 추가!
	t.Log("💤 running 상태가 될 때까지 대기")
	if err := waitUntilInstanceRunning(ctx, manager, user, 60*time.Second); err != nil {
		t.Fatalf("인스턴스 상태 대기 실패: %v", err)
	}

	// 2. 상태 확인
	t.Log("2️⃣ 상태 확인")
	status, err := manager.GetVMStatusByUser(ctx, user)
	if err != nil {
		t.Fatalf("상태 조회 실패: %v", err)
	}
	t.Logf("✅ 현재 상태: %s", status)

	// 3. 인스턴스 정지
	t.Log("3️⃣ 인스턴스 정지 시도")
	if err := manager.StopVMByUser(ctx, user); err != nil {
		t.Fatalf("인스턴스 정지 실패: %v", err)
	}
	time.Sleep(10 * time.Second) // EC2 API에 반영될 때까지 대기

	status, _ = manager.GetVMStatusByUser(ctx, user)
	t.Logf("✅ 정지 후 상태: %s", status)

	// 4. 인스턴스 시작
	t.Log("4️⃣ 인스턴스 시작 시도")
	if err := manager.StartVMByUser(ctx, user); err != nil {
		t.Fatalf("인스턴스 시작 실패: %v", err)
	}
	time.Sleep(10 * time.Second)

	status, _ = manager.GetVMStatusByUser(ctx, user)
	t.Logf("✅ 시작 후 상태: %s", status)

	// 5. 인스턴스 종료
	t.Log("5️⃣ 인스턴스 종료 시도")
	if err := manager.TerminateVMByUser(ctx, user); err != nil {
		t.Fatalf("인스턴스 종료 실패: %v", err)
	}
	t.Log("✅ 인스턴스 종료 완료")
}

func waitUntilInstanceRunning(ctx context.Context, m *myaws.EC2Manager, username string, timeout time.Duration) error {
	start := time.Now()
	for {
		state, err := m.GetVMStatusByUser(ctx, username)
		if err != nil {
			return fmt.Errorf("상태 조회 실패: %w", err)
		}
		if state == "running" {
			return nil
		}
		if time.Since(start) > timeout {
			return fmt.Errorf("타임아웃: 인스턴스가 running 상태로 전환되지 않음 (현재: %s)", state)
		}
		time.Sleep(5 * time.Second)
	}
}
