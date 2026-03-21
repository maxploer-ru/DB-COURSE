package handlers

import (
	"ZVideo/internal/domain/usecase/auth"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	registerUC *auth.RegisterUserUseCase
	//loginUC    *auth.LoginUserUseCase
}

func NewAuthHandler(
	registerUC *auth.RegisterUserUseCase,
	// loginUC *auth.LoginUserUseCase,
) *AuthHandler {
	return &AuthHandler{
		registerUC: registerUC,
		//loginUC:    loginUC,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var cmd auth.RegisterUserCommand

	if err := c.ShouldBindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.registerUC.Execute(c.Request.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		case errors.Is(err, auth.ErrUsernameAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		case errors.Is(err, auth.ErrWeakPassword):
			c.JSON(http.StatusBadRequest, gin.H{"error": "password is too weak"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, result)
}
