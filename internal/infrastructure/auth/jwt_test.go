package auth_test

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/auth"
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/suite"
)

type JwtServiceTestSuite struct {
	suite.Suite
	service *auth.JwtService
}

func (s *JwtServiceTestSuite) SetupTest() {
	s.service = auth.NewJwtService("access_secret", "refresh_secret", time.Hour, time.Hour*24)
}

func (s *JwtServiceTestSuite) TestGenerateAccessToken_Positive_StateTransition() {
	ctx := context.Background()
	data := &domain.AccessTokenData{
		UserID:   1,
		UserName: "test_user",
		Role:     "user",
	}

	token, err := s.service.GenerateAccessToken(ctx, data)

	s.NoError(err)
	s.NotEmpty(token)
}

func (s *JwtServiceTestSuite) TestGenerateAccessToken_Negative_InvalidData_EquivalencePartitioning() {
	ctx := context.Background()

	token, err := s.service.GenerateAccessToken(ctx, nil)

	s.Error(err)
	s.Empty(token)
}

func (s *JwtServiceTestSuite) TestGenerateRefreshToken_Positive_StateTransition() {
	ctx := context.Background()
	userID := 1

	token, data, err := s.service.GenerateRefreshToken(ctx, userID)

	s.NoError(err)
	s.NotEmpty(token)
	s.NotNil(data)
	s.Equal(userID, data.UserID)
	s.NotEmpty(data.TokenID)
}

func (s *JwtServiceTestSuite) TestGenerateRefreshToken_Negative_InvalidUserID_BoundaryValueAnalysis() {
	ctx := context.Background()

	token, data, err := s.service.GenerateRefreshToken(ctx, 0)

	s.Error(err)
	s.Empty(token)
	s.Nil(data)
}

func (s *JwtServiceTestSuite) TestValidateAccessToken_Positive_Combinatorial() {
	ctx := context.Background()
	data := &domain.AccessTokenData{
		UserID:   1,
		UserName: "test",
		Role:     "user",
	}
	token, _ := s.service.GenerateAccessToken(ctx, data)

	claims, err := s.service.ValidateAccessToken(ctx, token)

	s.NoError(err)
	s.NotNil(claims)
	s.Equal(1, claims.UserID)
	s.Equal("test", claims.UserName)
}

func (s *JwtServiceTestSuite) TestValidateAccessToken_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	invalidToken := "invalid.token.string"

	claims, err := s.service.ValidateAccessToken(ctx, invalidToken)

	s.Error(err)
	s.Nil(claims)
}

func (s *JwtServiceTestSuite) TestValidateAccessToken_Negative_DifferentHMACAlgorithm_Combinatorial() {
	ctx := context.Background()
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"user_id":  1,
		"username": "test",
		"role":     "user",
		"exp":      time.Now().Add(time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte("access_secret"))
	s.Require().NoError(err)

	claims, err := s.service.ValidateAccessToken(ctx, signed)

	s.Error(err)
	s.Nil(claims)
}

func (s *JwtServiceTestSuite) TestValidateRefreshToken_Positive_Combinatorial() {
	ctx := context.Background()
	token, generatedData, _ := s.service.GenerateRefreshToken(ctx, 1)

	claims, err := s.service.ValidateRefreshToken(ctx, token)

	s.NoError(err)
	s.NotNil(claims)
	s.Equal(1, claims.UserID)
	s.Equal(generatedData.TokenID, claims.TokenID)
}

func (s *JwtServiceTestSuite) TestValidateRefreshToken_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	invalidToken := "invalid.token.string"

	claims, err := s.service.ValidateRefreshToken(ctx, invalidToken)

	s.Error(err)
	s.Nil(claims)
}

func TestJwtServiceSuite(t *testing.T) {
	suite.Run(t, new(JwtServiceTestSuite))
}
