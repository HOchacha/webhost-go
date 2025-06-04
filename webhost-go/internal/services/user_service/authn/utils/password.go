package utils

import "golang.org/x/crypto/bcrypt"

type PasswordLocker interface {
	Hash(password string) (string, error)
	Verify(hash, password string) bool
}

type BcryptLocker struct{}

func (b *BcryptLocker) Hash(pw string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pw), 12)
	return string(bytes), err
}

func (b *BcryptLocker) Verify(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}
