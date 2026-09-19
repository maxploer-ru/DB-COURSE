package main

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/auth"
	"ZVideo/internal/infrastructure/config"
	"ZVideo/internal/infrastructure/db/postgres"
	pgmodels "ZVideo/internal/infrastructure/db/postgres/models"
	applogger "ZVideo/internal/infrastructure/logger"
	"ZVideo/internal/infrastructure/storage"
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type seedOptions struct {
	Count    int
	VideoDir string
	Driver   string
	Seed     int64
}

type seedAssets struct {
	VideoFiles []string
	Rng        *rand.Rand
}

func main() {
	cfg := config.LoadConfig()
	appLogger, closeLog := applogger.NewConfigured(cfg.Logging)
	defer closeLog()

	if err := run(cfg, appLogger); err != nil {
		appLogger.ErrorContext(context.Background(), "seed process failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(cfg *config.Config, appLogger domain.Logger) error {
	opts := parseFlags()
	if opts.VideoDir == "" {
		return fmt.Errorf("-video-dir is required")
	}
	if opts.Count <= 0 {
		return fmt.Errorf("-count must be positive")
	}

	if cfg == nil {
		return fmt.Errorf("failed to load config")
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}
	if opts.Driver != "" {
		cfg.DatabaseDriver = opts.Driver
	}

	videoFiles, err := collectVideoFiles(opts.VideoDir)
	if err != nil {
		return fmt.Errorf("collect video files: %w", err)
	}

	gofakeit.Seed(opts.Seed)
	assets := seedAssets{
		VideoFiles: videoFiles,
		Rng:        rand.New(rand.NewSource(opts.Seed)),
	}

	minioClient, _, err := storage.NewMinioClient(cfg.Minio)
	if err != nil {
		return fmt.Errorf("connect to MinIO: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := storage.EnsureBucketExists(ctx, minioClient, cfg.Minio.Bucket); err != nil {
		return fmt.Errorf("ensure MinIO bucket: %w", err)
	}

	switch strings.ToLower(cfg.DatabaseDriver) {
	case "postgres", "pg":
		if err := seedPostgres(cfg, minioClient, assets, opts.Count); err != nil {
			return fmt.Errorf("seed postgres: %w", err)
		}
	default:
		return fmt.Errorf("unsupported DB_DRIVER: %s", cfg.DatabaseDriver)
	}
	appLogger.InfoContext(context.Background(), "database seed completed", slog.Int("count", opts.Count), slog.String("driver", cfg.DatabaseDriver))
	return nil
}

func parseFlags() seedOptions {
	var opts seedOptions
	flag.IntVar(&opts.Count, "count", 1000, "records per table")
	flag.StringVar(&opts.VideoDir, "video-dir", "", "path to directory with .mp4 files")
	flag.Int64Var(&opts.Seed, "seed", time.Now().UnixNano(), "random seed")
	flag.Parse()
	return opts
}

func collectVideoFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if ext == ".mp4" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no .mp4 files found in %s", root)
	}
	return files, nil
}

func seedPostgres(cfg *config.Config, minioClient *minio.Client, assets seedAssets, count int) error {
	pgDB, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		return err
	}
	sqlDB, err := pgDB.DB()
	if err != nil {
		return fmt.Errorf("get postgres sql connection: %w", err)
	}
	defer func() { _ = sqlDB.Close() }()

	defaultRole, err := ensurePostgresRoles(pgDB)
	if err != nil {
		return err
	}

	pwdSvc := auth.NewBcryptPasswordService(0)
	passwordHash, err := pwdSvc.HashPassword(context.Background(), "Password123!")
	if err != nil {
		return err
	}

	const batchSize = 1000 // безопасный размер пачки для любого числа полей

	users := make([]pgmodels.User, 0, count)
	for i := 0; i < count; i++ {
		username := uniqueName("user", i)
		email := fmt.Sprintf("%s_%d@example.com", username, i)
		users = append(users, pgmodels.User{
			RoleID:               defaultRole.ID,
			Username:             username,
			Email:                email,
			PasswordHash:         passwordHash,
			IsActive:             true,
			NotificationsEnabled: assets.Rng.Intn(10) != 0,
			CreatedAt:            randomTime(assets.Rng),
			UpdatedAt:            randomTime(assets.Rng),
		})
	}
	if err := pgDB.CreateInBatches(users, batchSize).Error; err != nil {
		return err
	}

	channels := make([]pgmodels.Channel, 0, count)
	channelOwner := make(map[int]int, count)
	for i := 0; i < count; i++ {
		channel := pgmodels.Channel{
			UserID:      users[i].ID,
			Name:        fmt.Sprintf("%s_channel", users[i].Username),
			Description: gofakeit.Sentence(10),
			CreatedAt:   randomTime(assets.Rng),
		}
		channels = append(channels, channel)
	}
	if err := pgDB.CreateInBatches(channels, batchSize).Error; err != nil {
		return err
	}
	for i := range channels {
		channelOwner[channels[i].ID] = channels[i].UserID
	}

	videos := make([]pgmodels.Video, 0, count)
	for i := 0; i < count; i++ {
		channel := channels[assets.Rng.Intn(len(channels))]
		fileKey, err := uploadRandomVideo(minioClient, cfg.Minio.Bucket, assets, channel.ID)
		if err != nil {
			return err
		}
		videos = append(videos, pgmodels.Video{
			ChannelID:   channel.ID,
			Title:       gofakeit.Sentence(4),
			Description: gofakeit.Paragraph(1, 3, 10, " "),
			Filepath:    fileKey,
			CreatedAt:   randomTime(assets.Rng),
		})
	}
	if err := pgDB.CreateInBatches(videos, batchSize).Error; err != nil {
		return err
	}

	playlists := make([]pgmodels.Playlist, 0, count)
	for i := 0; i < count; i++ {
		channel := channels[assets.Rng.Intn(len(channels))]
		playlists = append(playlists, pgmodels.Playlist{
			ChannelID:   channel.ID,
			Name:        gofakeit.BuzzWord() + " Mix",
			Description: gofakeit.Sentence(8),
			CreatedAt:   randomTime(assets.Rng),
		})
	}
	if err := pgDB.CreateInBatches(playlists, batchSize).Error; err != nil {
		return err
	}

	playlistItemTarget := min(count, len(playlists)*len(videos))
	playlistItems := make([]pgmodels.PlaylistItem, 0, playlistItemTarget)
	usedPlaylistPairs := map[string]struct{}{}
	for len(playlistItems) < playlistItemTarget {
		playlist := playlists[assets.Rng.Intn(len(playlists))]
		video := videos[assets.Rng.Intn(len(videos))]
		key := fmt.Sprintf("%d:%d", playlist.ID, video.ID)
		if _, ok := usedPlaylistPairs[key]; ok {
			continue
		}
		usedPlaylistPairs[key] = struct{}{}
		playlistItems = append(playlistItems, pgmodels.PlaylistItem{
			PlaylistID: playlist.ID,
			VideoID:    video.ID,
			Number:     1,
			AddedAt:    randomTime(assets.Rng),
		})
	}
	if err := pgDB.CreateInBatches(playlistItems, batchSize).Error; err != nil {
		return err
	}

	viewings := make([]pgmodels.Viewing, 0, count*65)
	for i := 0; i < count*65; i++ {
		user := users[assets.Rng.Intn(len(users))]
		video := videos[assets.Rng.Intn(len(videos))]
		viewings = append(viewings, pgmodels.Viewing{
			UserID:    user.ID,
			VideoID:   video.ID,
			WatchedAt: randomTime(assets.Rng),
		})
	}
	if err := pgDB.CreateInBatches(viewings, batchSize).Error; err != nil {
		return err
	}

	comments := make([]pgmodels.Comment, 0, count*65)
	for i := 0; i < count*65; i++ {
		user := users[assets.Rng.Intn(len(users))]
		video := videos[assets.Rng.Intn(len(videos))]
		comments = append(comments, pgmodels.Comment{
			UserID:    user.ID,
			VideoID:   video.ID,
			Content:   gofakeit.Sentence(12),
			CreatedAt: randomTime(assets.Rng),
		})
	}
	if err := pgDB.CreateInBatches(comments, batchSize).Error; err != nil {
		return err
	}

	videoRatingTarget := min(count*65, len(users)*len(videos))
	videoRatings := make([]pgmodels.VideoRating, 0, videoRatingTarget)
	usedVideoPairs := map[string]struct{}{}
	for len(videoRatings) < videoRatingTarget {
		user := users[assets.Rng.Intn(len(users))]
		video := videos[assets.Rng.Intn(len(videos))]
		key := fmt.Sprintf("%d:%d", user.ID, video.ID)
		if _, ok := usedVideoPairs[key]; ok {
			continue
		}
		usedVideoPairs[key] = struct{}{}
		videoRatings = append(videoRatings, pgmodels.VideoRating{
			UserID:  user.ID,
			VideoID: video.ID,
			Liked:   assets.Rng.Intn(2) == 0,
			RatedAt: randomTime(assets.Rng),
		})
	}
	if err := pgDB.CreateInBatches(videoRatings, batchSize).Error; err != nil {
		return err
	}

	commentRatingTarget := min(count*65, len(users)*len(comments))
	commentRatings := make([]pgmodels.CommentRating, 0, commentRatingTarget)
	usedCommentPairs := map[string]struct{}{}
	for len(commentRatings) < commentRatingTarget {
		user := users[assets.Rng.Intn(len(users))]
		comment := comments[assets.Rng.Intn(len(comments))]
		key := fmt.Sprintf("%d:%d", user.ID, comment.ID)
		if _, ok := usedCommentPairs[key]; ok {
			continue
		}
		usedCommentPairs[key] = struct{}{}
		commentRatings = append(commentRatings, pgmodels.CommentRating{
			UserID:    user.ID,
			CommentID: comment.ID,
			Liked:     assets.Rng.Intn(2) == 0,
			RatedAt:   randomTime(assets.Rng),
		})
	}
	if err := pgDB.CreateInBatches(commentRatings, batchSize).Error; err != nil {
		return err
	}

	subscriptionTarget := min(count*10, len(users)*max(0, len(channelOwner)-1))
	subscriptions := make([]pgmodels.Subscription, 0, subscriptionTarget)
	usedSubPairs := map[string]struct{}{}
	for len(subscriptions) < subscriptionTarget {
		user := users[assets.Rng.Intn(len(users))]
		channelID := pickRandomChannel(channelOwner, assets)
		if channelOwner[channelID] == user.ID {
			continue
		}
		key := fmt.Sprintf("%d:%d", user.ID, channelID)
		if _, ok := usedSubPairs[key]; ok {
			continue
		}
		usedSubPairs[key] = struct{}{}
		subscriptions = append(subscriptions, pgmodels.Subscription{
			UserID:         user.ID,
			ChannelID:      channelID,
			NewVideosCount: assets.Rng.Intn(5),
			SubscribedAt:   randomTime(assets.Rng),
		})
	}
	if err := pgDB.CreateInBatches(subscriptions, batchSize).Error; err != nil {
		return err
	}

	return nil
}

func ensurePostgresRoles(db *gorm.DB) (*pgmodels.Role, error) {
	var defaultRole pgmodels.Role
	if err := db.Where("is_default = ?", true).First(&defaultRole).Error; err != nil {
		return nil, err
	}

	return &defaultRole, nil
}

func uploadRandomVideo(client *minio.Client, bucket string, assets seedAssets, channelID int) (string, error) {
	filePath := assets.VideoFiles[assets.Rng.Intn(len(assets.VideoFiles))]
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", err
	}

	key := fmt.Sprintf("videos/%d/%s_%s", channelID, uuid.New().String(), filepath.Base(filePath))
	_, err = client.PutObject(context.Background(), bucket, key, file, info.Size(), minio.PutObjectOptions{ContentType: "video/mp4"})
	if err != nil {
		return "", err
	}
	return key, nil
}

func randomTime(rng *rand.Rand) time.Time {
	end := time.Now().UTC()
	start := end.AddDate(-2, 0, 0)
	if end.Unix() <= start.Unix() {
		return end
	}
	delta := end.Unix() - start.Unix()
	return time.Unix(start.Unix()+rng.Int63n(delta), 0).UTC()
}

func pickRandomChannel(channelOwner map[int]int, assets seedAssets) int {
	idx := assets.Rng.Intn(len(channelOwner))
	pos := 0
	for channelID := range channelOwner {
		if pos == idx {
			return channelID
		}
		pos++
	}
	return 0
}

func uniqueName(prefix string, i int) string {
	suffix := strings.ReplaceAll(uuid.New().String(), "-", "")
	return fmt.Sprintf("%s_%d_%s", prefix, i, suffix[:8])
}
