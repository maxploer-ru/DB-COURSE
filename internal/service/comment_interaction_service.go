package service

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/repository"
	"context"
	"fmt"
	"log/slog"
)

type CommentInteractionService interface {
	Like(ctx context.Context, userID, commentID int) error
	Dislike(ctx context.Context, userID, commentID int) error
	RemoveRating(ctx context.Context, userID, commentID int) error
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

func (s *commentInteractionService) Like(ctx context.Context, userID, commentID int) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "CommentInteractionService"),
		slog.String("operation", "Like"),
		slog.Int("user_id", userID),
		slog.Int("comment_id", commentID),
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
	if existing != nil {
		if existing.Liked {
			logger.DebugContext(ctx, "Comment already liked, no change")
			return nil
		}
		existing.Liked = true
		logger.DebugContext(ctx, "Updating rating from dislike to like")
		if err := s.ratingRepo.Update(ctx, existing); err != nil {
			logger.ErrorContext(ctx, "Failed to update rating", slog.String("error", err.Error()))
			return err
		}
		_ = s.statsCache.IncrLikes(ctx, commentID)
		_ = s.statsCache.DecrDislikes(ctx, commentID)
		logger.DebugContext(ctx, "Rating updated to like")
		return nil
	}

	rating := &domain.CommentRating{
		UserID:    userID,
		CommentID: commentID,
		Liked:     true,
	}
	logger.DebugContext(ctx, "Creating new like rating")
	if err := s.ratingRepo.Create(ctx, rating); err != nil {
		logger.ErrorContext(ctx, "Failed to create rating", slog.String("error", err.Error()))
		return err
	}
	_ = s.statsCache.IncrLikes(ctx, commentID)
	logger.DebugContext(ctx, "Like rating created")
	return nil
}

func (s *commentInteractionService) Dislike(ctx context.Context, userID, commentID int) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "CommentInteractionService"),
		slog.String("operation", "Dislike"),
		slog.Int("user_id", userID),
		slog.Int("comment_id", commentID),
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
	if existing != nil {
		if !existing.Liked {
			logger.DebugContext(ctx, "Comment already disliked, no change")
			return nil
		}
		existing.Liked = false
		logger.DebugContext(ctx, "Updating rating from like to dislike")
		if err := s.ratingRepo.Update(ctx, existing); err != nil {
			logger.ErrorContext(ctx, "Failed to update rating", slog.String("error", err.Error()))
			return err
		}
		_ = s.statsCache.IncrDislikes(ctx, commentID)
		_ = s.statsCache.DecrLikes(ctx, commentID)
		logger.DebugContext(ctx, "Rating updated to dislike")
		return nil
	}

	rating := &domain.CommentRating{
		UserID:    userID,
		CommentID: commentID,
		Liked:     false,
	}
	logger.DebugContext(ctx, "Creating new dislike rating")
	if err := s.ratingRepo.Create(ctx, rating); err != nil {
		logger.ErrorContext(ctx, "Failed to create rating", slog.String("error", err.Error()))
		return err
	}
	_ = s.statsCache.IncrDislikes(ctx, commentID)
	logger.DebugContext(ctx, "Dislike rating created")
	return nil
}

func (s *commentInteractionService) RemoveRating(ctx context.Context, userID, commentID int) error {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "CommentInteractionService"),
		slog.String("operation", "RemoveRating"),
		slog.Int("user_id", userID),
		slog.Int("comment_id", commentID),
	)

	logger.DebugContext(ctx, "Checking existing rating")
	existing, err := s.ratingRepo.GetByUserAndComment(ctx, userID, commentID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get existing rating", slog.String("error", err.Error()))
		return err
	}
	if existing == nil {
		logger.WarnContext(ctx, "Rating not found")
		return domain.ErrCommentRatingNotFound
	}

	logger.DebugContext(ctx, "Deleting rating", slog.Bool("was_liked", existing.Liked))
	if err := s.ratingRepo.Delete(ctx, userID, commentID); err != nil {
		logger.ErrorContext(ctx, "Failed to delete rating", slog.String("error", err.Error()))
		return err
	}
	if existing.Liked {
		_ = s.statsCache.DecrLikes(ctx, commentID)
	} else {
		_ = s.statsCache.DecrDislikes(ctx, commentID)
	}
	logger.DebugContext(ctx, "Rating removed successfully")
	return nil
}

func (s *commentInteractionService) GetStats(ctx context.Context, commentID int) (likes, dislikes int64, err error) {
	logger := domain.GetLogger(ctx).With(
		slog.String("service", "CommentInteractionService"),
		slog.String("operation", "GetStats"),
		slog.Int("comment_id", commentID),
	)

	logger.DebugContext(ctx, "Trying to get stats from cache")
	likes, dislikes, hit, err := s.statsCache.GetStats(ctx, commentID)
	if err == nil && hit {
		logger.DebugContext(ctx, "Stats retrieved from cache", slog.Int64("likes", likes), slog.Int64("dislikes", dislikes))
		return likes, dislikes, nil
	}
	if err != nil {
		logger.WarnContext(ctx, "Cache error, falling back to DB", slog.String("error", err.Error()))
	}

	logger.DebugContext(ctx, "Fetching stats from database")
	likes, dislikes, err = s.ratingRepo.GetStats(ctx, commentID)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to get stats from database", slog.String("error", err.Error()))
		return 0, 0, fmt.Errorf("get stats from db: %w", err)
	}
	_ = s.statsCache.SetStats(ctx, commentID, likes, dislikes)
	logger.DebugContext(ctx, "Stats retrieved from DB and cached", slog.Int64("likes", likes), slog.Int64("dislikes", dislikes))
	return likes, dislikes, nil
}
