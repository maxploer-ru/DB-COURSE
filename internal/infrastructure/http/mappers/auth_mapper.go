package mappers

import (
	"ZVideo/internal/domain/entity"
	"ZVideo/internal/domain/usecase/auth"
	"ZVideo/internal/infrastructure/http/dto"
	"time"
)

type AuthMapper struct{}

func NewAuthMapper() *AuthMapper {
	return &AuthMapper{}
}

func (m *AuthMapper) ToRegisterCommand(req *dto.RegisterRequest) auth.RegisterUserCommand {
	return auth.RegisterUserCommand{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}
}

//func (m *AuthMapper) ToLoginCommand(req *dto.LoginRequest) auth.LoginUserCommand {
//	return auth.LoginUserCommand{
//		Email:    req.Email,
//		Password: req.Password,
//	}
//}
//
//func (m *AuthMapper) ToRefreshTokenCommand(req *dto.RefreshTokenRequest) auth.RefreshTokenCommand {
//	return auth.RefreshTokenCommand{
//		RefreshToken: req.RefreshToken,
//	}
//}

func (m *AuthMapper) ToAuthResponse(user *entity.User, accessToken, refreshToken string) dto.AuthResponse {
	return dto.AuthResponse{
		User:         m.ToUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}

func (m *AuthMapper) ToUserResponse(user *entity.User) dto.UserResponse {
	role := ""
	if user.Role != nil {
		role = user.Role.Name
	}

	return dto.UserResponse{
		ID:                   user.ID,
		Username:             user.Username,
		Email:                user.Email,
		NotificationsEnabled: user.NotificationsEnabled,
		Role:                 role,
		CreatedAt:            user.CreatedAt.Format(time.RFC3339),
	}
}
