package aws_test

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"os"
	"testing"
	"time"
	myaws "webhost-go/webhost-go/pkg/aws" // <- 여기를 실제 모듈 경로로 수정하세요
)

func TestStartUbuntuVM(t *testing.T) {
	ctx := context.Background()

	// AWS Config 로드 (환경변수 또는 IAM 역할 기반)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Fatalf("AWS 설정 로딩 실패: %v", err)
	}

	manager := myaws.NewEC2Manager(cfg)

	// 고유한 이름 지정 (시간 기반)
	name := "test-instance-" + time.Now().Format("150405")
	userID := int64(9999)

	t.Logf("Start EC2 instance: %s", name)
	instanceID, publicIP, err := manager.StartUbuntuVM(ctx, name, userID)
	if err != nil {
		t.Fatalf("EC2 인스턴스 생성 실패: %v", err)
	}

	t.Logf("생성 완료: ID=%s, PublicIP=%s", instanceID, publicIP)

	// (선택) 테스트 후 바로 종료
	if os.Getenv("CLEANUP") == "true" {
		t.Logf("테스트 인스턴스 종료 시도: %s", instanceID)
		_, err := manager.Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
			InstanceIds: []string{instanceID},
		})
		if err != nil {
			t.Logf("종료 실패: %v", err)
		} else {
			t.Log("인스턴스 종료 완료")
		}
	}
}
