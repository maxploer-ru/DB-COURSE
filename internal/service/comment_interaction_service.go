package service

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/repository"
	"context"
	"fmt"
	"log/slog"
)

type CommentInteractionService interface {
	Rate(ctx context.Context, userID, commentID int, action domain.RatingAction) error
	GetStats(ctx context.Context, commentID int) (likes, dislikes int64, err error)
}

type commentInteractionService struct {
	ratingRepo  repository.CommentRatingRepository
	commentRepo repository.CommentRepository
	statsCache  repository.CommentStatsCache
}

func NewCommentInteractionService(
	ratingRepo repository.CommentRatingRepository,
	commentRepo repository.CommentRepository,
	statsCache repository.CommentStatsCache,
) CommentInteractionService {
	return &commentInteractionService{
		ratingRepo:  ratingRepo,
		commentRepo: commentRepo,
		statsCache:  statsCache,
	}
}

func (s *commentInteractionService) Rate(ctx context.Context, userID, commentID int, action domain.RatingAction) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "CommentInteractionService"),
		slog.String("operation", "Rate"),
		slog.Int("user_id", userID),
		slog.Int("comment_id", commentID),
		slog.String("action", string(action)),
	)

	logger.DebugContext(ctx, "Checking comment existence")
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get comment", slog.String("error", err.Error()))
		return err
	}
	if comment == nil {
		logger.WarnContext(ctx, "Comment not found")
		return domain.ErrCommentNotFound
	}

	logger.DebugContext(ctx, "Checking existing rating")
	existing, err := s.ratingRepo.GetByUserAndComment(ctx, userID, commentID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get existing rating", slog.String("error", err.Error()))
		return err
	}

	if action == domain.RatingActionRemove {
		if existing == nil {
			return domain.ErrCommentRatingNotFound
		}
		if err := s.ratingRepo.Delete(ctx, userID, commentID); err != nil {
			logger.ErrorContext(ctx, "Failed to delete rating", slog.String("error", err.Error()))
			return err
		}

		if existing.Liked {
			if err := s.statsCache.DecrLikes(ctx, commentID); err != nil {
				logger.WarnContext(ctx, "Failed to decrement likes cache", slog.String("error", err.Error()))
			}
		} else {
			if err := s.statsCache.DecrDislikes(ctx, commentID); err != nil {
				logger.WarnContext(ctx, "Failed to decrement dislikes cache", slog.String("error", err.Error()))
			}
		}
		logger.InfoContext(ctx, "Rating removed successfully")
		return nil
	}

	isLike := action == domain.RatingActionLike

	if existing != nil {
		if existing.Liked == isLike {
			logger.DebugContext(ctx, "Rating is already in the requested state, no change")
			return nil
		}

		existing.Liked = isLike
		if err := s.ratingRepo.Update(ctx, existing); err != nil {
			logger.ErrorContext(ctx, "Failed to update rating", slog.String("error", err.Error()))
			return err
		}

		if isLike {
			if err := s.statsCache.IncrLikes(ctx, commentID); err != nil {
				logger.WarnContext(ctx, "Cache error", slog.String("error", err.Error()))
			}
			if err := s.statsCache.DecrDislikes(ctx, commentID); err != nil {
				logger.WarnContext(ctx, "Cache error", slog.String("error", err.Error()))
			}
		} else {
			if err := s.statsCache.IncrDislikes(ctx, commentID); err != nil {
				logger.WarnContext(ctx, "Cache error", slog.String("error", err.Error()))
			}
			if err := s.statsCache.DecrLikes(ctx, commentID); err != nil {
				logger.WarnContext(ctx, "Cache error", slog.String("error", err.Error()))
			}
		}
		logger.InfoContext(ctx, "Rating updated successfully")
		return nil
	}

	rating := &domain.CommentRating{
		UserID:    userID,
		CommentID: commentID,
		Liked:     isLike,
	}
	if err := s.ratingRepo.Create(ctx, rating); err != nil {
		logger.ErrorContext(ctx, "Failed to create rating", slog.String("error", err.Error()))
		return err
	}

	if isLike {
		if err := s.statsCache.IncrLikes(ctx, commentID); err != nil {
			logger.WarnContext(ctx, "Cache error", slog.String("error", err.Error()))
		}
	} else {
		if err := s.statsCache.IncrDislikes(ctx, commentID); err != nil {
			logger.WarnContext(ctx, "Cache error", slog.String("error", err.Error()))
		}
	}
	logger.InfoContext(ctx, "Rating created successfully")
	return nil
}

func (s *commentInteractionService) GetStats(ctx context.Context, commentID int) (likes, dislikes int64, err error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "CommentInteractionService"),
		slog.String("operation", "GetStats"),
		slog.Int("comment_id", commentID),
	)

	logger.DebugContext(ctx, "Trying to get stats from cache")
	likes, dislikes, hit, cacheErr := s.statsCache.GetStats(ctx, commentID)
	if cacheErr == nil && hit {
		logger.DebugContext(ctx, "Stats retrieved from cache", slog.Int64("likes", likes), slog.Int64("dislikes", dislikes))
		return likes, dislikes, nil
	}
	if cacheErr != nil {
		logger.WarnContext(ctx, "Cache error, falling back to DB", slog.String("error", cacheErr.Error()))
	}

	logger.DebugContext(ctx, "Fetching stats from database")
	likes, dislikes, err = s.ratingRepo.GetStats(ctx, commentID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get stats from database", slog.String("error", err.Error()))
		return 0, 0, fmt.Errorf("get stats from db: %w", err)
	}
	if cacheErr == nil {
		_ = s.statsCache.SetStats(ctx, commentID, likes, dislikes)
	}

	logger.DebugContext(ctx, "Stats retrieved from DB and cached", slog.Int64("likes", likes), slog.Int64("dislikes", dislikes))
	return likes, dislikes, nil
}
