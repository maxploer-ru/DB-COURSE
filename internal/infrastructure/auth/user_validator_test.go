package auth_test

import (
	"ZVideo/internal/infrastructure/auth"
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type UserValidatorTestSuite struct {
	suite.Suite
	validator *auth.UserValidator
}

func (s *UserValidatorTestSuite) SetupTest() {
	s.validator = auth.NewUserValidator()
}

func (s *UserValidatorTestSuite) TestValidateNewUser_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	email := "user@example.com"
	nickname := "validUser"
	password := "StrongPass123!"

	err := s.validator.ValidateNewUser(ctx, email, nickname, password)

	s.NoError(err)
}

func (s *UserValidatorTestSuite) TestValidateNewUser_Negative_Email_EquivalencePartitioning() {
	ctx := context.Background()
	email := "invalid-email"
	nickname := "validUser"
	password := "StrongPass123!"

	err := s.validator.ValidateNewUser(ctx, email, nickname, password)

	s.Error(err)
}

func (s *UserValidatorTestSuite) TestValidateNewUser_Negative_Username_BoundaryValueAnalysis() {
	ctx := context.Background()
	email := "user@example.com"
	nickname := "u"
	password := "StrongPass123!"

	err := s.validator.ValidateNewUser(ctx, email, nickname, password)

	s.Error(err)
}

func (s *UserValidatorTestSuite) TestValidateNewUser_Negative_Password_EquivalencePartitioning() {
	ctx := context.Background()
	email := "user@example.com"
	nickname := "validUser"
	password := "123"

	err := s.validator.ValidateNewUser(ctx, email, nickname, password)

	s.Error(err)
}

func TestUserValidatorSuite(t *testing.T) {
	suite.Run(t, new(UserValidatorTestSuite))
}
