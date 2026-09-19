package auth_test

import (
	"ZVideo/internal/infrastructure/auth"
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PasswordServiceTestSuite struct {
	suite.Suite
	service *auth.BcryptPasswordService
}

func (s *PasswordServiceTestSuite) SetupTest() {
	s.service = auth.NewBcryptPasswordService(4)
}

func (s *PasswordServiceTestSuite) TestHashPassword_Positive_StateTransition() {
	ctx := context.Background()
	password := "strong_password123"

	hash, err := s.service.HashPassword(ctx, password)

	s.NoError(err)
	s.NotEmpty(hash)
	s.NotEqual(password, hash)
}

func (s *PasswordServiceTestSuite) TestHashPassword_Negative_BoundaryValueAnalysis() {
	ctx := context.Background()
	password := generateLongString(75)

	hash, err := s.service.HashPassword(ctx, password)

	s.Error(err)
	s.Empty(hash)
}

func (s *PasswordServiceTestSuite) TestComparePassword_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	password := "strong_password123"
	hash, _ := s.service.HashPassword(ctx, password)

	err := s.service.ComparePassword(ctx, password, hash)

	s.NoError(err)
}

func (s *PasswordServiceTestSuite) TestComparePassword_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	password := "strong_password123"
	hash, _ := s.service.HashPassword(ctx, password)

	err := s.service.ComparePassword(ctx, "wrong_password", hash)

	s.Error(err)
}

func generateLongString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}

func TestPasswordServiceSuite(t *testing.T) {
	suite.Run(t, new(PasswordServiceTestSuite))
}
