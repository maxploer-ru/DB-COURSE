package main

import (
	"log"

	"github.com/gin-gonic/gin"

	authService "ZVideo/internal/domain/auth/service"
	authUC "ZVideo/internal/domain/auth/usecase"

	"ZVideo/internal/infrastructure/config"
	"ZVideo/internal/infrastructure/db/postgres"
	postgresRepo "ZVideo/internal/infrastructure/db/postgres/repositories"
	redisDB "ZVideo/internal/infrastructure/db/redis"
	"ZVideo/internal/infrastructure/http/handlers"
	"ZVideo/internal/infrastructure/http/mappers"
)

func main() {
	cfg := config.LoadConfig()
	if cfg == nil {
		log.Fatal("Failed to load configuration")
	}

	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	redisClient, err := redisDB.NewClient(cfg.Redis)
	if err != nil {
		log.Printf("⚠️  Redis connection failed: %v", err)
	}

	tokenRepo := redisDB.NewTokenRepository(redisClient)
	userRepo := postgresRepo.NewUserRepository(db)
	roleRepo := postgresRepo.NewRoleRepository(db)

	passwordSvc := authService.NewPasswordService()
	jwtSvc := authService.NewJWTService(&cfg.JWT)
	userValSvc := authService.NewUserValidationService(userRepo, cfg.Auth.Password)

	registerUC := authUC.NewRegisterUserUseCase(
		userRepo,
		roleRepo,
		tokenRepo,
		passwordSvc,
		jwtSvc,
		userValSvc,
	)

	authMapper := mappers.NewAuthMapper()
	authHandler := handlers.NewAuthHandler(registerUC, authMapper)

	r := gin.Default()

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)

		}

		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
	}

	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
