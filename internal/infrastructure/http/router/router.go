package router

import (
	"ZVideo/internal/infrastructure/http/handlers"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ZVideo/internal/infrastructure/config"
	httpMiddleware "ZVideo/internal/infrastructure/http/middleware"
)

type Router struct {
	engine         *gin.Engine
	config         *config.HTTPConfig
	authMiddleware *httpMiddleware.AuthMiddleware

	authHandler *handlers.AuthHandler
}

func NewRouter(
	cfg *config.HTTPConfig,
	authMiddleware *httpMiddleware.AuthMiddleware,
	authHandler *handlers.AuthHandler,
) *Router {
	engine := gin.New()

	r := &Router{
		engine:         engine,
		config:         cfg,
		authMiddleware: authMiddleware,
		authHandler:    authHandler,
	}

	r.setupMiddleware()
	r.setupRoutes()

	return r
}

func (r *Router) setupMiddleware() {
	r.engine.Use(gin.Recovery())
	r.engine.Use(gin.Logger())

	r.engine.Use(cors.New(cors.Config{
		AllowOrigins:     r.config.AllowedOrigins,
		AllowMethods:     r.config.AllowedMethods,
		AllowHeaders:     r.config.AllowedHeaders,
		ExposeHeaders:    r.config.ExposedHeaders,
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.engine.Use(requestIDMiddleware())
}

func (r *Router) setupRoutes() {
	v1 := r.engine.Group("/api/v1")

	auth := v1.Group("/auth")
	{
		auth.POST("/register", r.authHandler.Register)
	}
}

func (r *Router) Handler() http.Handler {
	return r.engine
}

func rateLimitMiddleware(limit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func generateRequestID() string {
	return uuid.New().String()
}
