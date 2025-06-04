package user_service

// Model Definition

type User struct {
	ID       int64
	Email    string
	Password string
	Role     string
	Name     string
}

// Role Enumerations

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)
