package auth

import (
	"ZVideo/internal/domain"
	"context"
	"fmt"
	"net/mail"
	"unicode"
)

type UserValidator struct{}

func NewUserValidator() *UserValidator {
	return &UserValidator{}
}

func (v *UserValidator) ValidateNewUser(ctx context.Context, email, nickname, password string) error {
	_ = ctx
	if err := v.validateNickname(nickname); err != nil {
		return fmt.Errorf("%w: %s", domain.ErrInvalidUsername, err.Error())
	}
	if err := v.validateEmail(email); err != nil {
		return fmt.Errorf("%w: %s", domain.ErrInvalidUserEmail, err.Error())
	}
	if err := v.validatePassword(password); err != nil {
		return fmt.Errorf("%w: %s", domain.ErrWeakPassword, err.Error())
	}
	return nil
}

func (v *UserValidator) validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email format: %w", err)
	}
	return nil
}

func (v *UserValidator) validateNickname(nickname string) error {
	if nickname == "" {
		return fmt.Errorf("nickname is required")
	}
	if len(nickname) < 3 {
		return fmt.Errorf("nickname must be at least 3 characters")
	}
	if len(nickname) > 32 {
		return fmt.Errorf("nickname must be at most 32 characters")
	}

	for _, ch := range nickname {
		switch {
		case unicode.IsLetter(ch):
			continue
		case unicode.IsDigit(ch):
			continue
		case ch == '_' || ch == '-' || ch == '.':
			continue
		default:
			return fmt.Errorf("nickname contains invalid character '%c' (allowed: letters, digits, _, -, .)", ch)
		}
	}
	return nil
}

func (v *UserValidator) validatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}
