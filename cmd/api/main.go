package main

import (
	authUC "ZVideo/internal/domain/usecase/auth"
	"ZVideo/internal/infrastructure/config"
	"ZVideo/internal/infrastructure/db/postgres"
	"ZVideo/internal/infrastructure/http/handlers"
	//"ZVideo/internal/infrastructure/http/router"
	"ZVideo/internal/pkg/jwt"
	"ZVideo/internal/pkg/password"
	"log"

	"github.com/gin-gonic/gin"

	postgresRepo "ZVideo/internal/infrastructure/db/postgres/repositories"
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

	userRepo := postgresRepo.NewUserRepository(db)
	roleRepo := postgresRepo.NewRoleRepository(db)

	passwordSvc := password.NewService()
	jwtSvc := jwt.NewService(&cfg.JWT)

	registerUC := authUC.NewRegisterUserUseCase(userRepo, roleRepo, passwordSvc, jwtSvc)
	//loginUC := authUC.NewLoginUserUseCase(userRepo, passwordSvc, jwtSvc)

	authHandler := handlers.NewAuthHandler(registerUC /*, loginUC*/)

	r := gin.Default()

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			//auth.POST("/login", authHandler.Login)
		}
	}

	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
