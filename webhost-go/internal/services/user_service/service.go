package user_service

type Service interface {
	// Authentication
	Signup(email, password, name string) error // User Create를 대신함
	Login(email, password string) (token string, err error)
	VerifyToken(token string) (*User, error)

	// User managerment
	GetUserByEmail(email string) (*User, error)
	ListUsers() ([]*User, error)
	UpdateUser(email, name, password string) error
	DeleteUser(id int64) error
}
