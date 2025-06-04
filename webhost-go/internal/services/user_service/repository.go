package user_service

type Repository interface {
	FindByID(id int64) (*User, error)
	FindByEmail(email string) (*User, error)
	FindAll() ([]*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(id int64) error
}
