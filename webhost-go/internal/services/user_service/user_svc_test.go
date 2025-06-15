package user_service_test

import (
	"fmt"
	"testing"
	"time"
	"webhost-go/webhost-go/internal/services/user_service"
	"webhost-go/webhost-go/internal/services/user_service/authn/token"
	"webhost-go/webhost-go/internal/services/user_service/authn/utils"

	"github.com/stretchr/testify/assert"
)

// 임시 테스트 구현체
type mockRepo struct {
	users map[string]*user_service.User
	idSeq int64
}

func newMockRepo() *mockRepo {
	return &mockRepo{users: make(map[string]*user_service.User)}
}

func (m *mockRepo) FindByEmail(email string) (*user_service.User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, fmt.Errorf("User not found")
	}
	return u, nil
}

func (m *mockRepo) Create(u *user_service.User) error {
	m.idSeq++
	u.ID = m.idSeq
	m.users[u.Email] = u
	return nil
}

func (m *mockRepo) Update(u *user_service.User) error {
	m.users[u.Email] = u
	return nil
}

func (m *mockRepo) Delete(id int64) error {
	for email, u := range m.users {
		if u.ID == id {
			delete(m.users, email)
			return nil
		}
	}
	return fmt.Errorf("User not found")
}

func (m *mockRepo) FindAll() ([]*user_service.User, error) {
	var all []*user_service.User
	for _, u := range m.users {
		all = append(all, u)
	}
	return all, nil
}

func (m *mockRepo) FindByID(id int64) (*user_service.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

// --- 테스트 시작 ---

func TestUserService(t *testing.T) {
	repo := newMockRepo()
	locker := &utils.BcryptLocker{}
	tokens := token.NewJWTManager("test-secret", 30*time.Minute)
	svc := user_service.NewService(repo, locker, tokens)

	email := "test@example.com"
	password := "password123"
	name := "Test User"

	// 1. 회원가입
	err := svc.Signup(email, password, name)
	assert.NoError(t, err)

	// 2. 중복 회원가입
	err = svc.Signup(email, password, name)
	assert.Error(t, err)

	// 3. 로그인
	tokenStr, err := svc.Login(email, password)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	// 4. 로그인 실패 (잘못된 비밀번호)
	_, err = svc.Login(email, "wrongpass")
	assert.Error(t, err)

	// 5. 토큰 검증
	validation := tokens.Validate(tokenStr)
	assert.True(t, validation.Valid)
	assert.Equal(t, email, validation.Claims.Email)

	// 6. 사용자 정보 업데이트
	newName := "Updated Name"
	newPassword := "newpass123"
	err = svc.UpdateUser(email, newName, newPassword)
	assert.NoError(t, err)

	// 7. 업데이트 후 로그인 확인
	tokenStr2, err := svc.Login(email, newPassword)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr2)

	// 7-1. 이름 변경 확인
	updatedUser, err := repo.FindByEmail(email)
	assert.NoError(t, err)
	assert.Equal(t, newName, updatedUser.Name)

	// 8. 사용자 목록
	users, err := svc.ListUsers()
	assert.NoError(t, err)
	assert.Len(t, users, 1)

	// 9. 사용자 삭제
	err = svc.DeleteUserByEmail(email)
	assert.NoError(t, err)

	// 9-1. 삭제 후 존재 확인
	_, err = repo.FindByEmail(email)
	assert.Error(t, err)

	// 10. 삭제 후 로그인 실패
	_, err = svc.Login(email, newPassword)
	assert.Error(t, err)
}
