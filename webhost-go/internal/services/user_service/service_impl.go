package user_service

import (
	"errors"
	"fmt"
	"webhost-go/webhost-go/internal/services/user_service/authn/token"
	"webhost-go/webhost-go/internal/services/user_service/authn/utils"
)

type service struct {
	repo   Repository
	locker utils.PasswordLocker
	tokens token.TokenManager
}

func NewService(r Repository, l utils.PasswordLocker, t token.TokenManager) Service {
	return &service{repo: r, locker: l, tokens: t}
}

func (s *service) Signup(email, password, name string) error {

	if _, err := s.repo.FindByEmail(email); err == nil {
		return errors.New("user already exists")
	}

	hashed, err := s.locker.Hash(password)
	if err != nil {
		return fmt.Errorf("fail to hash password: %w", err)
	}

	user := &User{
		Email:    email,
		Password: hashed,
		Name:     name,
		Role:     token.RoleUser, // 기본 역할
	}

	return s.repo.Create(user)
}

func (s *service) Login(email, password string) (string, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return "", fmt.Errorf("사용자 조회 실패: %w", err)
	}

	if !s.locker.Verify(user.Password, password) {
		return "", errors.New("잘못된 비밀번호입니다")
	}

	return s.tokens.Generate(user.Email, user.Role)
}

// ID는 말 그대로 숫자임, 아무래도 email 기반으로 찾게끔 해야 할 것 같아
func (s *service) UpdateUser(email, name, newPassword string) error {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return err
	}

	if name != "" {
		user.Name = name
	}
	if newPassword != "" {
		hashed, err := s.locker.Hash(newPassword)
		if err != nil {
			return err
		}
		user.Password = hashed
	}

	return s.repo.Update(user)
}

func (s *service) ListUsers() ([]*User, error) {
	return s.repo.FindAll()
}

func (s *service) DeleteUser(id int64) error {
	return s.repo.Delete(id)
}

func (s *service) DeleteUserByEmail(email string) error {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return err
	}
	return s.repo.Delete(user.ID)
}

func (s *service) GetUserByEmail(email string) (*User, error) {
	return s.repo.FindByEmail(email)
}
