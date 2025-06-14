package user_service

import "webhost-go/webhost-go/internal/services/user_service/authn/token"

// Model Definition

type User struct {
	ID       int64
	Email    string
	Password string
	Role     token.Role
	Name     string
}
