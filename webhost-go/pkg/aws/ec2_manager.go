package aws

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type EC2Manager struct {
	Client *ec2.Client
}

func NewEC2Manager(cfg aws.Config) *EC2Manager {
	return &EC2Manager{
		Client: ec2.NewFromConfig(cfg),
	}
}

func (m *EC2Manager) StartUbuntuVM(ctx context.Context, name string, userID int64) (instanceID string, publicIP string, err error) {
	// 1. 최신 Ubuntu AMI 검색
	amiID, err := FindLatestUbuntuAmi(ctx, m.Client)
	if err != nil {
		return "", "", fmt.Errorf("failed to find Ubuntu AMI: %w", err)
	}

	// 2. EC2 인스턴스 생성 요청
	input := &ec2.RunInstancesInput{
		ImageId:      aws.String(amiID),
		InstanceType: ec2types.InstanceTypeT2Micro,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeInstance,
				Tags: []ec2types.Tag{
					{Key: aws.String("Name"), Value: aws.String(name)},
					{Key: aws.String("Owner"), Value: aws.String(fmt.Sprint(userID))},
				},
			},
		},
	}

	runOut, err := m.Client.RunInstances(ctx, input)
	if err != nil || len(runOut.Instances) == 0 {
		return "", "", fmt.Errorf("EC2 run failed: %w", err)
	}

	instance := runOut.Instances[0]
	instanceID = *instance.InstanceId

	// 3. 퍼블릭 IP 확보 (잠시 대기 후 조회)
	time.Sleep(10 * time.Second)

	desc, err := m.Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return "", "", fmt.Errorf("describe failed: %w", err)
	}

	publicIP = aws.ToString(desc.Reservations[0].Instances[0].PublicIpAddress)
	return instanceID, publicIP, nil
}

// Ubuntu 공식 제공자 Canonical의 AWS 계정 ID
const CanonicalOwnerID = "099720109477"

// FindLatestUbuntuAmi searches for the latest Ubuntu 22.04 AMI in the given region
func FindLatestUbuntuAmi(ctx context.Context, ec2Client *ec2.Client) (string, error) {
	output, err := ec2Client.DescribeImages(ctx, &ec2.DescribeImagesInput{
		Owners: []string{CanonicalOwnerID},
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("name"),
				Values: []string{"ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"},
			},
			{
				Name:   aws.String("architecture"),
				Values: []string{"x86_64"},
			},
			{
				Name:   aws.String("root-device-type"),
				Values: []string{"ebs"},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to describe images: %w", err)
	}

	if len(output.Images) == 0 {
		return "", fmt.Errorf("no Ubuntu 22.04 images found")
	}

	// 최신 순 정렬
	sort.Slice(output.Images, func(i, j int) bool {
		return *output.Images[i].CreationDate > *output.Images[j].CreationDate
	})

	return *output.Images[0].ImageId, nil
}

func (m *EC2Manager) StartUbuntuVMWithOwner(ctx context.Context, name string, username string) (instanceID, publicIP string, priavteIP string, err error) {
	amiID, err := FindLatestUbuntuAmi(ctx, m.Client)
	if err != nil {
		return "", "", "", fmt.Errorf("AMI 검색 실패: %w", err)
	}

	input := &ec2.RunInstancesInput{
		ImageId:      aws.String(amiID),
		InstanceType: ec2types.InstanceTypeT2Micro,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeInstance,
				Tags: []ec2types.Tag{
					{Key: aws.String("Name"), Value: aws.String(name)},
					{Key: aws.String("User"), Value: aws.String(username)},
					{Key: aws.String("Role"), Value: aws.String("webhosting")},
				},
			},
		},
	}
	out, err := m.Client.RunInstances(ctx, input)
	if err != nil {
		return "", "", "", fmt.Errorf("EC2 생성 실패: %w", err)
	}

	instanceID = *out.Instances[0].InstanceId

	// 퍼블릭 IP 확보
	time.Sleep(10 * time.Second)
	desc, err := m.Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return "", "", "", fmt.Errorf("IP 조회 실패: %w", err)
	}
	publicIP = aws.ToString(desc.Reservations[0].Instances[0].PublicIpAddress)
	privateIP := aws.ToString(desc.Reservations[0].Instances[0].PrivateIpAddress)
	return instanceID, publicIP, privateIP, nil
}

func (m *EC2Manager) findInstanceByUser(ctx context.Context, username string) (*ec2types.Instance, error) {
	out, err := m.Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("tag:User"),
				Values: []string{username},
			},
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"pending", "running", "stopped"},
			},
			{
				Name:   aws.String("tag:Role"),
				Values: []string{"webhosting"},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	for _, r := range out.Reservations {
		for _, inst := range r.Instances {
			return &inst, nil
		}
	}
	return nil, fmt.Errorf("사용자 %s의 VM을 찾을 수 없음", username)
}

func (m *EC2Manager) GetVMStatusByUser(ctx context.Context, username string) (string, error) {
	inst, err := m.findInstanceByUser(ctx, username)
	if err != nil {
		return "", err
	}
	return string(inst.State.Name), nil
}

func (m *EC2Manager) StopVMByUser(ctx context.Context, username string) error {
	inst, err := m.findInstanceByUser(ctx, username)
	if err != nil {
		return fmt.Errorf("인스턴스 조회 실패: %w", err)
	}

	if err := isWebhostingAndNotProtected(inst); err != nil {
		return err
	}

	_, err = m.Client.StopInstances(ctx, &ec2.StopInstancesInput{
		InstanceIds: []string{aws.ToString(inst.InstanceId)},
	})
	if err != nil {
		return fmt.Errorf("인스턴스 중지 실패: %w", err)
	}

	return nil
}

func (m *EC2Manager) StartVMByUser(ctx context.Context, username string) error {
	inst, err := m.findInstanceByUser(ctx, username)
	if err != nil {
		return fmt.Errorf("인스턴스 조회 실패: %w", err)
	}

	if err := isWebhostingAndNotProtected(inst); err != nil {
		return err
	}

	_, err = m.Client.StartInstances(ctx, &ec2.StartInstancesInput{
		InstanceIds: []string{aws.ToString(inst.InstanceId)},
	})
	if err != nil {
		return fmt.Errorf("인스턴스 시작 실패: %w", err)
	}

	return nil
}

func (m *EC2Manager) TerminateVMByUser(ctx context.Context, username string) error {
	// 사용자 태그 기반으로 인스턴스 조회
	inst, err := m.findInstanceByUser(ctx, username)
	if err != nil {
		return fmt.Errorf("인스턴스 조회 실패: %w", err)
	}

	instanceID := aws.ToString(inst.InstanceId)

	// 태그 검사: 보호 인스턴스인지 확인
	var role, protected string
	for _, tag := range inst.Tags {
		switch aws.ToString(tag.Key) {
		case "Role":
			role = aws.ToString(tag.Value)
		case "Protected":
			protected = aws.ToString(tag.Value)
		}
	}

	// 1. 보호된 인스턴스는 종료 금지
	if protected == "true" {
		return fmt.Errorf("이 인스턴스는 Protected=true 태그가 설정되어 있어 종료할 수 없습니다 (ID: %s)", instanceID)
	}

	// 2. 웹호스팅(Role=webhosting)이 아닌 인스턴스도 종료 금지
	if role != "webhosting" {
		return fmt.Errorf("이 인스턴스는 Role=webhosting 이 아니므로 종료 대상이 아닙니다 (ID: %s, Role: %s)", instanceID, role)
	}

	// 3. 안전 확인 후 종료
	_, err = m.Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return fmt.Errorf("종료 요청 실패: %w", err)
	}

	return nil
}

func (m *EC2Manager) FindInstanceByUser(ctx context.Context, username string) (*ec2types.Instance, error) {
	out, err := m.Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("tag:User"),
				Values: []string{username},
			},
			{
				Name:   aws.String("tag:Role"),
				Values: []string{"webhosting"},
			},
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"pending", "running", "stopped"},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("DescribeInstances 실패: %w", err)
	}

	for _, r := range out.Reservations {
		for _, inst := range r.Instances {
			return &inst, nil
		}
	}

	return nil, fmt.Errorf("사용자 %s의 인스턴스를 찾을 수 없습니다", username)
}

func isWebhostingAndNotProtected(inst *ec2types.Instance) error {
	var role, protected string
	for _, tag := range inst.Tags {
		switch aws.ToString(tag.Key) {
		case "Role":
			role = aws.ToString(tag.Value)
		case "Protected":
			protected = aws.ToString(tag.Value)
		}
	}

	instanceID := aws.ToString(inst.InstanceId)

	if protected == "true" {
		return fmt.Errorf("이 인스턴스는 Protected=true 태그가 설정되어 있어 작업할 수 없습니다 (ID: %s)", instanceID)
	}

	if role != "webhosting" {
		return fmt.Errorf("이 인스턴스는 Role=webhosting 이 아니므로 작업 대상이 아닙니다 (ID: %s, Role: %s)", instanceID, role)
	}

	return nil
}
