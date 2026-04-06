package middleware

import (
	"github.com/gin-gonic/gin"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
)

func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	id, ok := userID.(string)
	return id, ok
}
