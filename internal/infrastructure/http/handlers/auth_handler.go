package handlers

import (
	"ZVideo/internal/domain/auth/service"
	"ZVideo/internal/domain/auth/usecase"
	"ZVideo/internal/infrastructure/http/dto"
	"ZVideo/internal/infrastructure/http/mappers"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	registerUC *usecase.RegisterUserUseCase
	//loginUC    *auth.LoginUserUseCase
	mapper *mappers.AuthMapper
}

func NewAuthHandler(
	registerUC *usecase.RegisterUserUseCase,
	// loginUC *auth.LoginUserUseCase,
	mapper *mappers.AuthMapper,
) *AuthHandler {
	return &AuthHandler{
		registerUC: registerUC,
		//loginUC:    loginUC,
		mapper: mapper,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	cmd := h.mapper.ToRegisterCommand(&req)

	result, err := h.registerUC.Execute(c.Request.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "email already exists"})
		case errors.Is(err, service.ErrUsernameAlreadyExists):
			c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "username already exists"})
		case errors.Is(err, service.ErrWeakPassword):
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "password is too weak"})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	response := h.mapper.ToAuthResponse(result.User, result.AccessToken, result.RefreshToken)
	c.JSON(http.StatusCreated, response)
}
