package main

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/infrastructure/auth"
	"ZVideo/internal/infrastructure/config"
	"ZVideo/internal/infrastructure/db/mongo"
	"ZVideo/internal/infrastructure/db/mongo/repository"
	"ZVideo/internal/infrastructure/db/postgres"
	pgmodels "ZVideo/internal/infrastructure/db/postgres/models"
	"ZVideo/internal/infrastructure/storage"
	context "context"
	"flag"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	opts := parseFlags()
	if opts.VideoDir == "" {
		panic("-video-dir is required")
	}

	cfg := config.LoadConfig()
	if cfg == nil {
		panic("failed to load config")
	}
	if opts.Driver != "" {
		cfg.DatabaseDriver = opts.Driver
	}

	videoFiles, err := collectVideoFiles(opts.VideoDir)
	if err != nil {
		panic(err)
	}

	gofakeit.Seed(opts.Seed)
	assets := seedAssets{
		VideoFiles: videoFiles,
		Rng:        rand.New(rand.NewSource(opts.Seed)),
	}

	minioClient, _, err := storage.NewMinioClient(cfg.Minio)
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := storage.EnsureBucketExists(ctx, minioClient, cfg.Minio.Bucket); err != nil {
		cancel()
		panic(err)
	}
	cancel()

	switch strings.ToLower(cfg.DatabaseDriver) {
	case "mongo", "mongodb":
		if err := seedMongo(cfg, minioClient, assets, opts.Count); err != nil {
			panic(err)
		}
	case "postgres", "pg":
		if err := seedPostgres(cfg, minioClient, assets, opts.Count); err != nil {
			panic(err)
		}
	default:
		panic("unsupported DB_DRIVER: " + cfg.DatabaseDriver)
	}
}

func parseFlags() seedOptions {
	var opts seedOptions
	flag.IntVar(&opts.Count, "count", 1000, "records per table")
	flag.StringVar(&opts.VideoDir, "video-dir", "", "path to directory with .mp4 files")
	flag.StringVar(&opts.Driver, "driver", "", "override DB_DRIVER (mongo/postgres)")
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

func seedMongo(cfg *config.Config, minioClient *minio.Client, assets seedAssets, count int) error {
	mongoConn, err := mongo.NewConnection(cfg.Mongo)
	if err != nil {
		return err
	}
	defer func() {
		_ = mongoConn.Close(context.Background())
	}()

	roleRepo := repository.NewRoleRepository(mongoConn.DB)
	userRepo := repository.NewUserRepository(mongoConn.DB)
	channelRepo := repository.NewChannelRepository(mongoConn.DB)
	videoRepo := repository.NewVideoRepository(mongoConn.DB)
	playlistRepo := repository.NewPlaylistRepository(mongoConn.DB)
	viewingRepo := repository.NewViewingRepository(mongoConn.DB)
	commentRepo := repository.NewCommentRepository(mongoConn.DB)
	videoRatingRepo := repository.NewVideoRatingRepository(mongoConn.DB)
	commentRatingRepo := repository.NewCommentRatingRepository(mongoConn.DB)
	subscriptionRepo := repository.NewSubscriptionRepository(mongoConn.DB)

	defaultRole, err := ensureMongoRoles(roleRepo, count)
	if err != nil {
		return err
	}

	pwdSvc := auth.NewBcryptPasswordService(0)
	passwordHash, err := pwdSvc.HashPassword(context.Background(), "Password123!")
	if err != nil {
		return err
	}

	users := make([]*domain.User, 0, count)
	for i := 0; i < count; i++ {
		username := uniqueName("user", i)
		email := fmt.Sprintf("%s_%d@example.com", username, i)
		user := &domain.User{
			Username:             username,
			Email:                email,
			PasswordHash:         passwordHash,
			IsActive:             true,
			NotificationsEnabled: assets.Rng.Intn(10) != 0,
			Role:                 defaultRole,
		}
		if err := userRepo.Create(context.Background(), user); err != nil {
			return err
		}
		users = append(users, user)
	}

	channels := make([]*domain.Channel, 0, count)
	channelOwner := make(map[int]int, count)
	for i := 0; i < count; i++ {
		owner := users[i]
		channel := &domain.Channel{
			UserID:      owner.ID,
			Name:        fmt.Sprintf("%s_channel", owner.Username),
			Description: gofakeit.Sentence(10),
			CreatedAt:   randomTime(assets.Rng),
		}
		if err := channelRepo.Create(context.Background(), channel); err != nil {
			return err
		}
		channels = append(channels, channel)
		channelOwner[channel.ID] = owner.ID
	}

	videos := make([]*domain.Video, 0, count)
	for i := 0; i < count; i++ {
		channel := channels[assets.Rng.Intn(len(channels))]
		fileKey, err := uploadRandomVideo(minioClient, cfg.Minio.Bucket, assets, channel.ID)
		if err != nil {
			return err
		}
		video := &domain.Video{
			ChannelID:   channel.ID,
			Title:       gofakeit.Sentence(4),
			Description: gofakeit.Paragraph(1, 3, 10, " "),
			Filepath:    fileKey,
			CreatedAt:   randomTime(assets.Rng),
		}
		if err := videoRepo.Create(context.Background(), video); err != nil {
			return err
		}
		videos = append(videos, video)
	}

	playlists := make([]*domain.Playlist, 0, count)
	for i := 0; i < count; i++ {
		channel := channels[assets.Rng.Intn(len(channels))]
		playlist := &domain.Playlist{
			ChannelID:   channel.ID,
			Name:        gofakeit.BuzzWord() + " Mix",
			Description: gofakeit.Sentence(8),
			CreatedAt:   randomTime(assets.Rng),
		}
		if err := playlistRepo.Create(context.Background(), playlist); err != nil {
			return err
		}
		playlists = append(playlists, playlist)
	}
	for i := 0; i < count; i++ {
		playlist := playlists[i]
		video := videos[assets.Rng.Intn(len(videos))]
		if err := playlistRepo.AddVideo(context.Background(), playlist.ID, video.ID); err != nil {
			return err
		}
	}

	comments := make([]*domain.Comment, 0, count)
	for i := 0; i < count; i++ {
		user := users[assets.Rng.Intn(len(users))]
		video := videos[assets.Rng.Intn(len(videos))]
		comment := &domain.Comment{
			UserID:    user.ID,
			VideoID:   video.ID,
			Content:   gofakeit.Sentence(12),
			CreatedAt: randomTime(assets.Rng),
		}
		if err := commentRepo.Create(context.Background(), comment); err != nil {
			return err
		}
		comments = append(comments, comment)
	}

	if err := seedMongoViewings(viewingRepo, users, videos, assets, count); err != nil {
		return err
	}
	if err := seedMongoVideoRatings(videoRatingRepo, users, videos, assets, count); err != nil {
		return err
	}
	if err := seedMongoCommentRatings(commentRatingRepo, users, comments, assets, count); err != nil {
		return err
	}
	if err := seedMongoSubscriptions(subscriptionRepo, channelOwner, users, assets, count); err != nil {
		return err
	}

	return nil
}

func ensureMongoRoles(roleRepo *repository.RoleRepository, count int) (*domain.Role, error) {
	ctx := context.Background()
	defaultRole, err := roleRepo.GetDefaultRole(ctx)
	if err != nil {
		return nil, err
	}
	if defaultRole == nil {
		role := &domain.Role{Name: "user", IsDefault: true}
		if err := roleRepo.Create(ctx, role); err != nil {
			return nil, err
		}
		defaultRole = role
	}

	seedRoles := []domain.Role{{Name: "admin"}, {Name: "moderator"}}
	for _, role := range seedRoles {
		r := role
		r.IsDefault = false
		_ = roleRepo.Create(ctx, &r)
	}

	for i := 0; i < max(0, count-3); i++ {
		role := &domain.Role{Name: uniqueName("role", i)}
		if err := roleRepo.Create(ctx, role); err != nil {
			return nil, err
		}
	}
	return defaultRole, nil
}

func seedMongoViewings(viewingRepo *repository.ViewingRepository, users []*domain.User, videos []*domain.Video, assets seedAssets, count int) error {
	for i := 0; i < count; i++ {
		user := users[assets.Rng.Intn(len(users))]
		video := videos[assets.Rng.Intn(len(videos))]
		viewing := &domain.Viewing{
			UserID:    user.ID,
			VideoID:   video.ID,
			WatchedAt: randomTime(assets.Rng),
		}
		if err := viewingRepo.Create(context.Background(), viewing); err != nil {
			return err
		}
	}
	return nil
}

func seedMongoVideoRatings(videoRatingRepo *repository.VideoRatingRepository, users []*domain.User, videos []*domain.Video, assets seedAssets, count int) error {
	used := make(map[string]struct{}, count)
	for len(used) < count {
		user := users[assets.Rng.Intn(len(users))]
		video := videos[assets.Rng.Intn(len(videos))]
		key := fmt.Sprintf("%d:%d", user.ID, video.ID)
		if _, ok := used[key]; ok {
			continue
		}
		used[key] = struct{}{}
		rating := &domain.VideoRating{UserID: user.ID, VideoID: video.ID, Liked: assets.Rng.Intn(2) == 0}
		if err := videoRatingRepo.Create(context.Background(), rating); err != nil {
			return err
		}
	}
	return nil
}

func seedMongoCommentRatings(commentRatingRepo *repository.CommentRatingRepository, users []*domain.User, comments []*domain.Comment, assets seedAssets, count int) error {
	used := make(map[string]struct{}, count)
	for len(used) < count {
		user := users[assets.Rng.Intn(len(users))]
		comment := comments[assets.Rng.Intn(len(comments))]
		key := fmt.Sprintf("%d:%d", user.ID, comment.ID)
		if _, ok := used[key]; ok {
			continue
		}
		used[key] = struct{}{}
		rating := &domain.CommentRating{UserID: user.ID, CommentID: comment.ID, Liked: assets.Rng.Intn(2) == 0}
		if err := commentRatingRepo.Create(context.Background(), rating); err != nil {
			return err
		}
	}
	return nil
}

func seedMongoSubscriptions(subscriptionRepo *repository.SubscriptionRepository, channelOwner map[int]int, users []*domain.User, assets seedAssets, count int) error {
	used := make(map[string]struct{}, count)
	for len(used) < count {
		user := users[assets.Rng.Intn(len(users))]
		channelID := pickRandomChannel(channelOwner, assets)
		if channelOwner[channelID] == user.ID {
			continue
		}
		key := fmt.Sprintf("%d:%d", user.ID, channelID)
		if _, ok := used[key]; ok {
			continue
		}
		used[key] = struct{}{}
		if _, err := subscriptionRepo.Subscribe(context.Background(), user.ID, channelID); err != nil {
			return err
		}
	}
	return nil
}

func seedPostgres(cfg *config.Config, minioClient *minio.Client, assets seedAssets, count int) error {
	pgDB, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		return err
	}

	defaultRole, err := ensurePostgresRoles(pgDB, count)
	if err != nil {
		return err
	}

	pwdSvc := auth.NewBcryptPasswordService(0)
	passwordHash, err := pwdSvc.HashPassword(context.Background(), "Password123!")
	if err != nil {
		return err
	}

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
	if err := pgDB.Create(&users).Error; err != nil {
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
	if err := pgDB.Create(&channels).Error; err != nil {
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
	if err := pgDB.Create(&videos).Error; err != nil {
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
	if err := pgDB.Create(&playlists).Error; err != nil {
		return err
	}

	playlistItems := make([]pgmodels.PlaylistItem, 0, count)
	usedPlaylistPairs := map[string]struct{}{}
	for len(playlistItems) < count {
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
	if err := pgDB.Create(&playlistItems).Error; err != nil {
		return err
	}

	viewings := make([]pgmodels.Viewing, 0, count)
	for i := 0; i < count; i++ {
		user := users[assets.Rng.Intn(len(users))]
		video := videos[assets.Rng.Intn(len(videos))]
		viewings = append(viewings, pgmodels.Viewing{
			UserID:    user.ID,
			VideoID:   video.ID,
			WatchedAt: randomTime(assets.Rng),
		})
	}
	if err := pgDB.Create(&viewings).Error; err != nil {
		return err
	}

	comments := make([]pgmodels.Comment, 0, count)
	for i := 0; i < count; i++ {
		user := users[assets.Rng.Intn(len(users))]
		video := videos[assets.Rng.Intn(len(videos))]
		comments = append(comments, pgmodels.Comment{
			UserID:    user.ID,
			VideoID:   video.ID,
			Content:   gofakeit.Sentence(12),
			CreatedAt: randomTime(assets.Rng),
		})
	}
	if err := pgDB.Create(&comments).Error; err != nil {
		return err
	}

	videoRatings := make([]pgmodels.VideoRating, 0, count)
	usedVideoPairs := map[string]struct{}{}
	for len(videoRatings) < count {
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
	if err := pgDB.Create(&videoRatings).Error; err != nil {
		return err
	}

	commentRatings := make([]pgmodels.CommentRating, 0, count)
	usedCommentPairs := map[string]struct{}{}
	for len(commentRatings) < count {
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
	if err := pgDB.Create(&commentRatings).Error; err != nil {
		return err
	}

	subscriptions := make([]pgmodels.Subscription, 0, count)
	usedSubPairs := map[string]struct{}{}
	for len(subscriptions) < count {
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
	if err := pgDB.Create(&subscriptions).Error; err != nil {
		return err
	}

	return nil
}

func ensurePostgresRoles(db *gorm.DB, count int) (*pgmodels.Role, error) {
	seedRoles := []pgmodels.Role{{Name: "admin"}, {Name: "moderator"}, {Name: "user", IsDefault: true}}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&seedRoles).Error; err != nil {
		return nil, err
	}

	var defaultRole pgmodels.Role
	if err := db.Where("is_default = ?", true).First(&defaultRole).Error; err != nil {
		return nil, err
	}

	extraRoles := make([]pgmodels.Role, 0, max(0, count-3))
	for i := 0; i < count-3; i++ {
		extraRoles = append(extraRoles, pgmodels.Role{Name: uniqueName("role", i)})
	}
	if len(extraRoles) > 0 {
		if err := db.Create(&extraRoles).Error; err != nil {
			return nil, err
		}
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
