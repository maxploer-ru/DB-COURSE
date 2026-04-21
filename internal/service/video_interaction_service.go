package service

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/repository"
	"context"
	"fmt"
	"log/slog"
)

type VideoInteractionService interface {
	Like(ctx context.Context, userID, videoID int) error
	Dislike(ctx context.Context, userID, videoID int) error
	RemoveRating(ctx context.Context, userID, videoID int) error
	RecordView(ctx context.Context, userID, videoID int) error
	GetStats(ctx context.Context, videoID int) (stats *domain.VideoStats, err error)
}

type videoInteractionService struct {
	ratingRepo  repository.VideoRatingRepository
	viewingRepo repository.ViewingRepository
	videoRepo   repository.VideoRepository
	commentRepo repository.CommentRepository
	statsCache  repository.VideoStatsCache
}

func NewVideoInteractionService(
	ratingRepo repository.VideoRatingRepository,
	viewingRepo repository.ViewingRepository,
	videoRepo repository.VideoRepository,
	commentRepo repository.CommentRepository,
	statsCache repository.VideoStatsCache,
) VideoInteractionService {
	return &videoInteractionService{
		ratingRepo:  ratingRepo,
		viewingRepo: viewingRepo,
		videoRepo:   videoRepo,
		commentRepo: commentRepo,
		statsCache:  statsCache,
	}
}

func (s *videoInteractionService) Like(ctx context.Context, userID, videoID int) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "VideoInteractionService"),
		slog.String("operation", "Like"),
		slog.Int("user_id", userID),
		slog.Int("video_id", videoID),
	)

	logger.DebugContext(ctx, "Checking video existence")
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get video", slog.String("error", err.Error()))
		return err
	}
	if video == nil {
		logger.WarnContext(ctx, "Video not found")
		return domain.ErrVideoNotFound
	}

	logger.DebugContext(ctx, "Checking existing rating")
	existing, err := s.ratingRepo.GetByUserAndVideo(ctx, userID, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get existing rating", slog.String("error", err.Error()))
		return err
	}
	if existing != nil {
		if existing.Liked {
			logger.DebugContext(ctx, "Video already liked, no change")
			return nil
		}
		existing.Liked = true
		logger.DebugContext(ctx, "Updating rating from dislike to like")
		if err := s.ratingRepo.Update(ctx, existing); err != nil {
			logger.ErrorContext(ctx, "Failed to update rating", slog.String("error", err.Error()))
			return err
		}
		_ = s.statsCache.IncrLikes(ctx, videoID)
		_ = s.statsCache.DecrDislikes(ctx, videoID)
		logger.InfoContext(ctx, "Rating updated to like")
		return nil
	}

	rating := &domain.VideoRating{
		UserID:  userID,
		VideoID: videoID,
		Liked:   true,
	}
	logger.DebugContext(ctx, "Creating new like rating")
	if err := s.ratingRepo.Create(ctx, rating); err != nil {
		logger.ErrorContext(ctx, "Failed to create rating", slog.String("error", err.Error()))
		return err
	}
	_ = s.statsCache.IncrLikes(ctx, videoID)
	logger.InfoContext(ctx, "Like rating created")
	return nil
}

func (s *videoInteractionService) Dislike(ctx context.Context, userID, videoID int) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "VideoInteractionService"),
		slog.String("operation", "Dislike"),
		slog.Int("user_id", userID),
		slog.Int("video_id", videoID),
	)

	logger.DebugContext(ctx, "Checking video existence")
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get video", slog.String("error", err.Error()))
		return err
	}
	if video == nil {
		logger.WarnContext(ctx, "Video not found")
		return domain.ErrVideoNotFound
	}

	logger.DebugContext(ctx, "Checking existing rating")
	existing, err := s.ratingRepo.GetByUserAndVideo(ctx, userID, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get existing rating", slog.String("error", err.Error()))
		return err
	}
	if existing != nil {
		if !existing.Liked {
			logger.DebugContext(ctx, "Video already disliked, no change")
			return nil
		}
		existing.Liked = false
		logger.DebugContext(ctx, "Updating rating from like to dislike")
		if err := s.ratingRepo.Update(ctx, existing); err != nil {
			logger.ErrorContext(ctx, "Failed to update rating", slog.String("error", err.Error()))
			return err
		}
		_ = s.statsCache.IncrDislikes(ctx, videoID)
		_ = s.statsCache.DecrLikes(ctx, videoID)
		logger.InfoContext(ctx, "Rating updated to dislike")
		return nil
	}

	rating := &domain.VideoRating{
		UserID:  userID,
		VideoID: videoID,
		Liked:   false,
	}
	logger.DebugContext(ctx, "Creating new dislike rating")
	if err := s.ratingRepo.Create(ctx, rating); err != nil {
		logger.ErrorContext(ctx, "Failed to create rating", slog.String("error", err.Error()))
		return err
	}
	_ = s.statsCache.IncrDislikes(ctx, videoID)
	logger.InfoContext(ctx, "Dislike rating created")
	return nil
}

func (s *videoInteractionService) RemoveRating(ctx context.Context, userID, videoID int) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "VideoInteractionService"),
		slog.String("operation", "RemoveRating"),
		slog.Int("user_id", userID),
		slog.Int("video_id", videoID),
	)

	logger.DebugContext(ctx, "Checking existing rating")
	existing, err := s.ratingRepo.GetByUserAndVideo(ctx, userID, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get existing rating", slog.String("error", err.Error()))
		return err
	}
	if existing == nil {
		logger.WarnContext(ctx, "Rating not found")
		return domain.ErrRatingNotFound
	}

	logger.DebugContext(ctx, "Deleting rating", slog.Bool("was_liked", existing.Liked))
	if err := s.ratingRepo.Delete(ctx, userID, videoID); err != nil {
		logger.ErrorContext(ctx, "Failed to delete rating", slog.String("error", err.Error()))
		return err
	}
	if existing.Liked {
		_ = s.statsCache.DecrLikes(ctx, videoID)
	} else {
		_ = s.statsCache.DecrDislikes(ctx, videoID)
	}
	logger.InfoContext(ctx, "Rating removed")
	return nil
}

func (s *videoInteractionService) RecordView(ctx context.Context, userID, videoID int) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "VideoInteractionService"),
		slog.String("operation", "RecordView"),
		slog.Int("user_id", userID),
		slog.Int("video_id", videoID),
	)

	logger.DebugContext(ctx, "Recording view")
	viewing := &domain.Viewing{
		UserID:  userID,
		VideoID: videoID,
	}
	if err := s.viewingRepo.Create(ctx, viewing); err != nil {
		logger.ErrorContext(ctx, "Failed to record view", slog.String("error", err.Error()))
		return fmt.Errorf("record view failed: %w", err)
	}
	_ = s.statsCache.IncrViews(ctx, videoID)
	logger.InfoContext(ctx, "View recorded")
	return nil
}

func (s *videoInteractionService) GetStats(ctx context.Context, videoID int) (stats *domain.VideoStats, err error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "VideoInteractionService"),
		slog.String("operation", "GetStats"),
		slog.Int("video_id", videoID),
	)

	logger.DebugContext(ctx, "Trying to get stats from cache")
	stats, hit, err := s.statsCache.GetStats(ctx, videoID)
	if err == nil && hit {
		logger.DebugContext(ctx, "Stats retrieved from cache",
			slog.Int("views", stats.Views),
			slog.Int("likes", stats.Likes),
			slog.Int("dislikes", stats.Dislikes),
			slog.Int("comments", stats.Comments))
		return stats, nil
	}
	if err != nil {
		logger.WarnContext(ctx, "Cache error, falling back to DB", slog.String("error", err.Error()))
	}

	logger.DebugContext(ctx, "Fetching stats from database")
	likes, dislikes, err := s.ratingRepo.GetStats(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get rating stats", slog.String("error", err.Error()))
		return &domain.VideoStats{}, err
	}
	views, err := s.viewingRepo.GetTotalViews(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get total views", slog.String("error", err.Error()))
		return &domain.VideoStats{}, err
	}
	comments, err := s.commentRepo.CountByVideo(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to count comments", slog.String("error", err.Error()))
		return &domain.VideoStats{}, err
	}

	logger.DebugContext(ctx, "Stats retrieved from DB and cache populated",
		slog.Int("views", views),
		slog.Int("likes", likes),
		slog.Int("dislikes", dislikes),
		slog.Int("comments", int(comments)))

	stats = &domain.VideoStats{
		Views:    views,
		Likes:    likes,
		Dislikes: dislikes,
		Comments: int(comments),
	}
	_ = s.statsCache.SetStats(ctx, videoID, stats)
	return stats, nil
}
