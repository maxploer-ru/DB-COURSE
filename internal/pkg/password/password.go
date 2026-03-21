package password

import (
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Hash(password string) (string, error)
	Verify(hashed, plain string) bool
	IsStrong(password string) bool
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *service) Verify(hashed, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
	return err == nil
}

func (s *service) IsStrong(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper, hasLower, hasDigit := false, false, false

	for _, ch := range password {
		switch {
		case 'A' <= ch && ch <= 'Z':
			hasUpper = true
		case 'a' <= ch && ch <= 'z':
			hasLower = true
		case '0' <= ch && ch <= '9':
			hasDigit = true
		}
	}

	return hasUpper && hasLower && hasDigit
}
