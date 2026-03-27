package service_test

import (
	"testing"
	"time"

	"ZVideo/internal/domain/auth/service"
	"ZVideo/internal/infrastructure/config"

	"github.com/stretchr/testify/assert"
)

func TestJWTService_GenerateAccessToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:          "test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: time.Hour * 24 * 7,
	}
	svc := service.NewJWTService(cfg)

	token, err := svc.GenerateAccessToken(1, "john", "user")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := svc.ValidateAccessToken(token)
	assert.NoError(t, err)
	assert.Equal(t, 1, claims.UserID)
	assert.Equal(t, "john", claims.Username)
	assert.Equal(t, "user", claims.Role)
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
}

func TestJWTService_GenerateRefreshToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:          "test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: time.Hour * 24 * 7,
	}
	svc := service.NewJWTService(cfg)

	token, err := svc.GenerateRefreshToken(1)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	userID, err := svc.ValidateRefreshToken(token)
	assert.NoError(t, err)
	assert.Equal(t, 1, userID)
}

func TestJWTService_ValidateAccessToken_Invalid(t *testing.T) {
	cfg := &config.JWTConfig{Secret: "test-secret"}
	svc := service.NewJWTService(cfg)

	_, err := svc.ValidateAccessToken("invalid")
	assert.Error(t, err)
}

func TestJWTService_ValidateRefreshToken_Invalid(t *testing.T) {
	cfg := &config.JWTConfig{Secret: "test-secret"}
	svc := service.NewJWTService(cfg)

	_, err := svc.ValidateRefreshToken("invalid")
	assert.Error(t, err)
}
