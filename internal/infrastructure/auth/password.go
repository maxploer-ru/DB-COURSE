package auth

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type BcryptPasswordService struct {
	cost int
}

func NewBcryptPasswordService(cost int) *BcryptPasswordService {
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptPasswordService{cost: cost}
}

func (s *BcryptPasswordService) HashPassword(ctx context.Context, password string) (string, error) {
	_ = ctx
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash failed: %w", err)
	}
	return string(hashedBytes), nil
}

func (s *BcryptPasswordService) ComparePassword(ctx context.Context, password, hash string) error {
	_ = ctx
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf("invalid password: %w", err)
	}
	return nil
}
