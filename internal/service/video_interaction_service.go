package service

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/repository"
	"context"
	"fmt"
	"log/slog"
)

type VideoInteractionService interface {
	Rate(ctx context.Context, userID, videoID int, action domain.RatingAction) error
	RecordView(ctx context.Context, userID, videoID int) error
	GetStats(ctx context.Context, videoID int) (*domain.VideoStats, error)
	GetStatsBatch(ctx context.Context, videoIDs []int) (map[int]*domain.VideoStats, error)
}

type videoInteractionService struct {
	ratingRepo  repository.VideoRatingRepository
	viewingRepo repository.ViewingRepository
	videoRepo   repository.VideoRepository
	commentRepo repository.CommentRepository
	statsCache  repository.VideoStatsCache
}

type atomicVideoRatingRepository interface {
	Upsert(ctx context.Context, rating *domain.VideoRating) (previous *domain.VideoRating, created bool, err error)
	DeleteAndGet(ctx context.Context, userID, videoID int) (previous *domain.VideoRating, found bool, err error)
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

func (s *videoInteractionService) Rate(ctx context.Context, userID, videoID int, action domain.RatingAction) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "VideoInteractionService"),
		slog.String("operation", "Rate"),
		slog.Int("user_id", userID),
		slog.Int("video_id", videoID),
		slog.String("action", string(action)),
	)
	if action != domain.RatingActionLike && action != domain.RatingActionDislike && action != domain.RatingActionRemove {
		return domain.ErrInvalidRatingAction
	}

	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get video", slog.String("error", err.Error()))
		return err
	}
	if video == nil || video.Status != domain.VideoStatusReady {
		return domain.ErrVideoNotFound
	}

	existing, err := s.ratingRepo.GetByUserAndVideo(ctx, userID, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get existing rating", slog.String("error", err.Error()))
		return err
	}

	if action == domain.RatingActionRemove {
		if atomicRepo, ok := s.ratingRepo.(atomicVideoRatingRepository); ok {
			previous, found, err := atomicRepo.DeleteAndGet(ctx, userID, videoID)
			if err != nil {
				return err
			}
			if !found {
				return domain.ErrRatingNotFound
			}
			if previous.Liked {
				_ = s.statsCache.DecrLikes(ctx, videoID)
			} else {
				_ = s.statsCache.DecrDislikes(ctx, videoID)
			}
			return nil
		}
		if existing == nil {
			return domain.ErrRatingNotFound
		}
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

	if atomicRepo, ok := s.ratingRepo.(atomicVideoRatingRepository); ok {
		rating := &domain.VideoRating{UserID: userID, VideoID: videoID, Liked: action == domain.RatingActionLike}
		previous, created, err := atomicRepo.Upsert(ctx, rating)
		if err != nil {
			return err
		}
		if created {
			if rating.Liked {
				_ = s.statsCache.IncrLikes(ctx, videoID)
			} else {
				_ = s.statsCache.IncrDislikes(ctx, videoID)
			}
		} else if previous.Liked != rating.Liked {
			if rating.Liked {
				_ = s.statsCache.IncrLikes(ctx, videoID)
				_ = s.statsCache.DecrDislikes(ctx, videoID)
			} else {
				_ = s.statsCache.IncrDislikes(ctx, videoID)
				_ = s.statsCache.DecrLikes(ctx, videoID)
			}
		}
		return nil
	}

	isLike := action == domain.RatingActionLike

	if existing != nil {
		if existing.Liked == isLike {
			return nil
		}

		existing.Liked = isLike
		if err := s.ratingRepo.Update(ctx, existing); err != nil {
			logger.ErrorContext(ctx, "Failed to update rating", slog.String("error", err.Error()))
			return err
		}

		if isLike {
			_ = s.statsCache.IncrLikes(ctx, videoID)
			_ = s.statsCache.DecrDislikes(ctx, videoID)
		} else {
			_ = s.statsCache.IncrDislikes(ctx, videoID)
			_ = s.statsCache.DecrLikes(ctx, videoID)
		}
		logger.InfoContext(ctx, "Rating updated")
		return nil
	}

	rating := &domain.VideoRating{
		UserID:  userID,
		VideoID: videoID,
		Liked:   isLike,
	}
	if err := s.ratingRepo.Create(ctx, rating); err != nil {
		logger.ErrorContext(ctx, "Failed to create rating", slog.String("error", err.Error()))
		return err
	}

	if isLike {
		_ = s.statsCache.IncrLikes(ctx, videoID)
	} else {
		_ = s.statsCache.IncrDislikes(ctx, videoID)
	}
	logger.InfoContext(ctx, "Rating created")
	return nil
}

func (s *videoInteractionService) RecordView(ctx context.Context, userID, videoID int) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "VideoInteractionService"),
		slog.String("operation", "RecordView"),
		slog.Int("user_id", userID),
		slog.Int("video_id", videoID),
	)

	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get video", slog.String("error", err.Error()))
		return err
	}
	if video == nil || video.Status != domain.VideoStatusReady {
		return domain.ErrVideoNotFound
	}

	viewing := &domain.Viewing{
		UserID:  userID,
		VideoID: videoID,
	}
	if err := s.viewingRepo.Create(ctx, viewing); err != nil {
		logger.ErrorContext(ctx, "Failed to record view", slog.String("error", err.Error()))
		return fmt.Errorf("record view failed: %w", err)
	}

	_ = s.statsCache.IncrViews(ctx, videoID)
	return nil
}

func (s *videoInteractionService) GetStats(ctx context.Context, videoID int) (*domain.VideoStats, error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "VideoInteractionService"),
		slog.String("operation", "GetStats"),
		slog.Int("video_id", videoID),
	)

	stats, hit, cacheErr := s.statsCache.GetStats(ctx, videoID)
	if cacheErr == nil && hit {
		return stats, nil
	}
	if cacheErr != nil {
		logger.WarnContext(ctx, "Cache error, falling back to DB", slog.String("error", cacheErr.Error()))
	}

	likes, dislikes, err := s.ratingRepo.GetStats(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get rating stats", slog.String("error", err.Error()))
		return nil, err
	}

	views, err := s.viewingRepo.GetTotalViews(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get total views", slog.String("error", err.Error()))
		return nil, err
	}

	comments, err := s.commentRepo.CountByVideo(ctx, videoID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to count comments", slog.String("error", err.Error()))
		return nil, err
	}

	stats = &domain.VideoStats{
		Views:    views,
		Likes:    likes,
		Dislikes: dislikes,
		Comments: int(comments),
	}

	if cacheErr == nil {
		_ = s.statsCache.SetStats(ctx, videoID, stats)
	}
	return stats, nil
}

func (s *videoInteractionService) GetStatsBatch(ctx context.Context, videoIDs []int) (map[int]*domain.VideoStats, error) {
	logger := domain.GetLogger(ctx).With(
		"service", "VideoInteractionService",
		"operation", "GetStatsBatch",
	)

	result := make(map[int]*domain.VideoStats, len(videoIDs))
	if len(videoIDs) == 0 {
		return result, nil
	}

	for _, id := range videoIDs {
		stats, err := s.GetStats(ctx, id)
		if err != nil {
			logger.WarnContext(ctx, "Failed to get stats for video in batch",
				"video_id", id,
				"error", err.Error(),
			)
			result[id] = &domain.VideoStats{}
			continue
		}
		result[id] = stats
	}

	return result, nil
}
