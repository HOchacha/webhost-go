package db_driver

import (
	"database/sql"
	"errors"
	"webhost-go/webhost-go/internal/services/user_service"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByID(id int64) (*user_service.User, error) {
	row := r.db.QueryRow("SELECT id, email, password, name FROM users WHERE id = ?", id)

	var u user_service.User
	if err := row.Scan(&u.ID, &u.Email, &u.Password, &u.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByEmail(email string) (*user_service.User, error) {
	row := r.db.QueryRow("SELECT id, email, password, name FROM users WHERE email = ?", email)

	var u user_service.User
	if err := row.Scan(&u.ID, &u.Email, &u.Password, &u.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindAll() ([]*user_service.User, error) {
	rows, err := r.db.Query("SELECT id, email, password, name FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*user_service.User
	for rows.Next() {
		var u user_service.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Password, &u.Name); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}

func (r *UserRepository) Create(u *user_service.User) error {
	_, err := r.db.Exec(
		"INSERT INTO users (email, password, name) VALUES (?, ?, ?)",
		u.Email, u.Password, u.Name,
	)
	return err
}

func (r *UserRepository) Update(u *user_service.User) error {
	_, err := r.db.Exec(
		"UPDATE users SET email = ?, password = ?, name = ? WHERE id = ?",
		u.Email, u.Password, u.Name, u.ID,
	)
	return err
}

func (r *UserRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}
