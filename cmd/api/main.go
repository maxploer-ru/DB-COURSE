package main

import (
	"ZVideo/internal/delivery/handlers"
	"ZVideo/internal/delivery/middleware"
	"ZVideo/internal/delivery/openapi"
	"ZVideo/internal/delivery/response"
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/auth"
	"ZVideo/internal/infrastructure/cache"
	"ZVideo/internal/infrastructure/config"
	"ZVideo/internal/infrastructure/db/postgres"
	pgrepository "ZVideo/internal/infrastructure/db/postgres/repository"
	"ZVideo/internal/infrastructure/logger"
	"ZVideo/internal/infrastructure/storage"
	"ZVideo/internal/service"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.LoadConfig()
	baseLogger, closeLog := logger.NewConfigured(cfg.Logging)
	defer closeLog()

	if err := run(cfg, baseLogger); err != nil {
		baseLogger.ErrorContext(context.Background(), "API process failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(cfg *config.Config, baseLogger domain.Logger) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	if !strings.EqualFold(cfg.DatabaseDriver, "postgres") && !strings.EqualFold(cfg.DatabaseDriver, "pg") {
		return fmt.Errorf("unsupported database driver %q", cfg.DatabaseDriver)
	}

	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get postgres sql connection: %w", err)
	}
	defer func() { _ = sqlDB.Close() }()

	minioClient, presignClient, err := storage.NewMinioClient(cfg.Minio)
	if err != nil {
		return fmt.Errorf("connect to MinIO: %w", err)
	}
	setupCtx, setupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer setupCancel()
	if err := storage.EnsureBucketExists(setupCtx, minioClient, cfg.Minio.Bucket); err != nil {
		return fmt.Errorf("ensure MinIO bucket: %w", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Database,
	})
	defer redisClient.Close()
	redisCtx, redisCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer redisCancel()
	if err := redisClient.Ping(redisCtx).Err(); err != nil {
		return fmt.Errorf("connect to Redis: %w", err)
	}

	userRepository := pgrepository.NewUserRepository(db)
	roleRepository := pgrepository.NewRoleRepository(db)
	channelRepository := pgrepository.NewChannelRepository(db)
	communityRepository := pgrepository.NewCommunityRepository(db)
	videoRepository := pgrepository.NewVideoRepository(db)
	subscriptionRepository := pgrepository.NewSubscriptionRepository(db)
	videoRatingRepository := pgrepository.NewVideoRatingRepository(db)
	viewingRepository := pgrepository.NewViewingRepository(db)
	commentRepository := pgrepository.NewCommentRepository(db)
	commentRatingRepository := pgrepository.NewCommentRatingRepository(db)
	playlistRepository := pgrepository.NewPlaylistRepository(db)

	subscriberCounter := cache.NewRedisSubscriberCounter(redisClient)
	videoStatsCache := cache.NewVideoStatsCache(redisClient)
	commentStatsCache := cache.NewCommentStatsCache(redisClient)
	refreshSessionCache := cache.NewRefreshSessionCache(redisClient)

	passwordService := auth.NewBcryptPasswordService(0)
	jwtService := auth.NewJwtService(cfg.JWT.Secret, cfg.JWT.Secret+"_refresh", cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)
	userValidator := auth.NewUserValidator()
	authService := service.NewAuthService(userRepository, roleRepository, refreshSessionCache, passwordService, jwtService, userValidator)
	channelService := service.NewChannelService(channelRepository)
	communityService := service.NewCommunityService(communityRepository, channelService)
	subscriptionService := service.NewSubscriptionService(subscriptionRepository, channelRepository, subscriberCounter)
	storageService := storage.NewMinioStorageService(minioClient, presignClient, cfg.Minio.Bucket)
	videoService := service.NewVideoService(videoRepository, subscriptionService, channelService, storageService)
	playlistService := service.NewPlaylistService(playlistRepository, videoRepository, channelService)
	videoInteractionService := service.NewVideoInteractionService(videoRatingRepository, viewingRepository, videoRepository, commentRepository, videoStatsCache)
	commentService := service.NewCommentService(commentRepository, videoRepository, videoStatsCache, channelService)
	commentInteractionService := service.NewCommentInteractionService(commentRatingRepository, commentRepository, commentStatsCache)
	userService := service.NewUserService(userRepository)
	adminService := service.NewAdminService(userRepository, roleRepository)
	handler := handlers.NewHandler(
		authService,
		adminService,
		userService,
		channelService,
		subscriptionService,
		videoService,
		videoInteractionService,
		commentService,
		commentInteractionService,
		playlistService,
		communityService,
	)

	router := chi.NewRouter()
	router.Use(middleware.Logging(baseLogger))
	router.Use(middleware.Recovery)
	apiHandler := openapi.HandlerWithOptions(handler, openapi.ChiServerOptions{
		// The generated paths already contain the /api/v1 version prefix.
		BaseURL:    "",
		BaseRouter: router,
		Middlewares: []openapi.MiddlewareFunc{
			selectiveAuthMiddleware(authService),
		},
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			domain.GetLogger(r.Context()).WarnContext(r.Context(), "OpenAPI request validation failed", slog.Any("error", err))
			response.RespondWithError(w, http.StatusBadRequest, "INVALID_PARAMETER", err.Error())
		},
	})

	server := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      apiHandler,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	serverErr := make(chan error, 1)
	go func() {
		baseLogger.InfoContext(ctx, "HTTP server started", slog.String("addr", server.Addr), slog.String("base_path", "/api/v1"))
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			serverErr <- nil
			return
		}
		serverErr <- err
	}()

	select {
	case <-ctx.Done():
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("HTTP server failed: %w", err)
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("HTTP server shutdown failed: %w", err)
	}
	baseLogger.InfoContext(context.Background(), "HTTP server stopped")
	return nil
}

func selectiveAuthMiddleware(authService service.AuthService) openapi.MiddlewareFunc {
	authMiddleware := middleware.Auth(authService)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, protected := r.Context().Value(openapi.BearerAuthScopes).([]string); protected {
				authMiddleware(next).ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
