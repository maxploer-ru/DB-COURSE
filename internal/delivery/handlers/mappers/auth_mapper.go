package mappers

import (
	"ZVideo/internal/delivery/handlers/dto"
	"ZVideo/internal/domain"
)

func ToAuthResponse(user *domain.User, accessToken, refreshToken string) dto.AuthResponse {
	return dto.AuthResponse{
		User:         ToUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}

func ToUserResponse(user *domain.User) dto.UserResponse {
	role := ""
	if user.Role != nil {
		role = user.Role.Name
	}

	return dto.UserResponse{
		ID:                   user.ID,
		Username:             user.Username,
		Email:                user.Email,
		Role:                 role,
		NotificationsEnabled: user.NotificationsEnabled,
	}
}

func ToValidateTokenResponse(data *domain.AccessTokenData) dto.ValidateTokenResponse {
	return dto.ValidateTokenResponse{
		UserID:   data.UserID,
		Username: data.UserName,
		Role:     data.Role,
	}
}
