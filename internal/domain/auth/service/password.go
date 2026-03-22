package service

import (
	"golang.org/x/crypto/bcrypt"
)

type PasswordService interface {
	Hash(password string) (string, error)

	Verify(hashed, plain string) bool
}

type passwordService struct {
}

func NewPasswordService() PasswordService {
	return &passwordService{}
}

func (s *passwordService) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *passwordService) Verify(hashed, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
	return err == nil
}
