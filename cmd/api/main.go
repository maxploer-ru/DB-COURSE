package main

import (
	"ZVideo/internal/delivery/handlers"
	"ZVideo/internal/delivery/router"
	"ZVideo/internal/infrastructure/auth"
	"ZVideo/internal/infrastructure/cache"
	"ZVideo/internal/infrastructure/config"
	"ZVideo/internal/infrastructure/db/postgres"
	"ZVideo/internal/infrastructure/db/postgres/repository"
	"ZVideo/internal/infrastructure/logger"
	"ZVideo/internal/infrastructure/storage"
	"ZVideo/internal/service"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	cfg := config.LoadConfig()
	if cfg == nil {
		log.Fatal("Failed to load configuration")
	}

	var logOutput io.Writer
	if cfg.Logging.OutputPath == "stdout" {
		logOutput = os.Stdout
	} else {
		logOutput = &lumberjack.Logger{
			Filename:   cfg.Logging.OutputPath,
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		}
	}
	level := parseLogLevel(cfg.Logging.Level)
	baseLogger := logger.NewSlogLogger(level, logOutput, cfg.Logging.AddSource)
	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	minioClient, err := storage.NewMinioClient(storage.MinioConfig{
		Endpoint:  cfg.Minio.Endpoint,
		AccessKey: cfg.Minio.AccessKey,
		SecretKey: cfg.Minio.SecretKey,
		UseSSL:    cfg.Minio.UseSSL,
		Bucket:    cfg.Minio.Bucket,
	})
	if err != nil {
		log.Fatal("MinIO client initialization failed:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := storage.EnsureBucketExists(ctx, minioClient, cfg.Minio.Bucket); err != nil {
		cancel()
		log.Fatal("MinIO bucket setup failed:", err)
	}
	cancel()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Database,
	})

	userRepository := repository.NewUserRepository(db)
	roleRepository := repository.NewRoleRepository(db)
	channelRepository := repository.NewChannelRepository(db)
	communityRepository := repository.NewCommunityRepository(db)
	videoRepository := repository.NewVideoRepository(db)
	subscriptionRepository := repository.NewSubscriptionRepository(db)
	videoRatingRepository := repository.NewVideoRatingRepository(db)
	viewingRepository := repository.NewViewingRepository(db)
	commentRepository := repository.NewCommentRepository(db)
	commentRatingRepository := repository.NewCommentRatingRepository(db)
	playlistRepository := repository.NewPlaylistRepository(db)

	counter := cache.NewRedisSubscriberCounter(redisClient)
	statsCache := cache.NewVideoStatsCache(redisClient)
	commentStatsCache := cache.NewCommentStatsCache(redisClient)
	refreshSessionCache := cache.NewRefreshSessionCache(redisClient)

	passwordService := auth.NewBcryptPasswordService(0)
	jwtService := auth.NewJwtService(cfg.JWT.Secret, cfg.JWT.Secret+"_refresh", cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)
	userValidationService := auth.NewUserValidator()
	authService := service.NewAuthService(userRepository, roleRepository, refreshSessionCache, passwordService, jwtService, userValidationService)
	storageService := storage.NewMinioStorageService(minioClient, cfg.Minio.Bucket)
	channelService := service.NewChannelService(channelRepository, videoRepository, storageService)
	communityService := service.NewCommunityService(communityRepository, channelService, userRepository)
	videoService := service.NewVideoService(videoRepository, subscriptionRepository, channelService, storageService)
	subscriptionService := service.NewSubscriptionService(subscriptionRepository, channelRepository, counter)
	playlistService := service.NewPlaylistService(playlistRepository, videoRepository, channelService)
	interactionService := service.NewVideoInteractionService(videoRatingRepository, viewingRepository, videoRepository, commentRepository, statsCache)
	commentService := service.NewCommentService(commentRepository, videoRepository, statsCache, channelService)
	commentInteractionService := service.NewCommentInteractionService(commentRatingRepository, commentRepository, commentStatsCache)
	adminService := service.NewAdminService(userRepository)

	h := router.Handlers{
		Auth:               handlers.NewAuthHandler(authService),
		Channel:            handlers.NewChannelHandler(channelService, subscriptionService),
		CommunityHandler:   handlers.NewCommunityHandler(communityService, subscriptionService),
		Video:              handlers.NewVideoHandler(videoService, interactionService),
		Subscription:       handlers.NewSubscriptionHandler(subscriptionService, channelService),
		VideoInteraction:   handlers.NewVideoInteractionHandler(interactionService),
		Comment:            handlers.NewCommentHandler(commentService, commentInteractionService),
		CommentInteraction: handlers.NewCommentInteractionHandler(commentInteractionService),
		Admin:              handlers.NewAdminHandler(adminService),
		Playlist:           handlers.NewPlaylistHandler(playlistService),
	}
	r := router.NewRouter(&h, authService, baseLogger)

	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := http.ListenAndServe(":"+cfg.Server.Port, r); err != nil {
		log.Fatal("Server failed:", err)
	}
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
