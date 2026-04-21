package auth

import (
	"context"
	"fmt"
	"time"

	"ZVideo/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JwtService struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewJwtService(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *JwtService {
	return &JwtService{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

type accessClaims struct {
	UserID   int    `json:"user_id"`
	UserName string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type refreshClaims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func (s *JwtService) GenerateAccessToken(ctx context.Context, data *domain.AccessTokenData) (string, error) {
	_ = ctx
	now := time.Now()
	claims := accessClaims{
		UserID:   data.UserID,
		UserName: data.UserName,
		Role:     data.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.accessSecret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

func (s *JwtService) GenerateRefreshToken(ctx context.Context, userID int) (string, *domain.RefreshTokenData, error) {
	_ = ctx
	now := time.Now()
	tokenID := uuid.NewString()
	expiresAt := now.Add(s.refreshTTL)
	claims := refreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.refreshSecret)
	if err != nil {
		return "", nil, fmt.Errorf("sign refresh token: %w", err)
	}
	return signed, &domain.RefreshTokenData{
		UserID:    userID,
		TokenID:   tokenID,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *JwtService) ValidateAccessToken(ctx context.Context, tokenString string) (*domain.AccessTokenData, error) {
	_ = ctx
	claims := &accessClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.accessSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse access token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid access token")
	}
	return &domain.AccessTokenData{
		UserID:   claims.UserID,
		UserName: claims.UserName,
		Role:     claims.Role,
	}, nil
}

func (s *JwtService) ValidateRefreshToken(ctx context.Context, tokenString string) (*domain.RefreshTokenData, error) {
	_ = ctx
	claims := &refreshClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.refreshSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse refresh token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}
	if claims.ID == "" || claims.ExpiresAt == nil {
		return nil, fmt.Errorf("invalid refresh token claims")
	}
	return &domain.RefreshTokenData{
		UserID:    claims.UserID,
		TokenID:   claims.ID,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}
