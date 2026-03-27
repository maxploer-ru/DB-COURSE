package service

import (
	"ZVideo/internal/domain/video/entity"
	"ZVideo/internal/domain/video/repository"
	"context"
	"errors"
	"time"
)

type VideoRatingService interface {
	RateVideo(ctx context.Context, userID, videoID int, liked bool) error
	GetUserRating(ctx context.Context, userID, videoID int) (*entity.VideoRating, error)
}

type videoRatingService struct {
	videoRepo  repository.VideoRepository
	ratingRepo repository.VideoRatingRepository
	statsCache repository.VideoStatsCache
}

func NewVideoRatingService(
	videoRepo repository.VideoRepository,
	ratingRepo repository.VideoRatingRepository,
	statsCache repository.VideoStatsCache,
) VideoRatingService {
	return &videoRatingService{
		videoRepo:  videoRepo,
		ratingRepo: ratingRepo,
		statsCache: statsCache,
	}
}

func (s *videoRatingService) RateVideo(ctx context.Context, userID, videoID int, liked bool) error {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return err
	}
	if video == nil {
		return ErrVideoNotFound
	}

	existing, err := s.ratingRepo.GetByUserAndVideo(ctx, userID, videoID)
	if err != nil && !errors.Is(err, ErrRatingNotFound) {
		return err
	}

	if existing == nil {
		rating := &entity.VideoRating{
			UserID:  userID,
			VideoID: videoID,
			Liked:   liked,
			RatedAt: time.Now(),
		}
		if err := s.ratingRepo.Create(ctx, rating); err != nil {
			return err
		}
	} else {
		if existing.Liked == liked {
			return nil
		}

		existing.Liked = liked
		existing.RatedAt = time.Now()

		if err := s.ratingRepo.Update(ctx, existing); err != nil {
			return err
		}
	}

	_ = s.statsCache.Invalidate(ctx, videoID)
	return nil
}

func (s *videoRatingService) GetUserRating(ctx context.Context, userID, videoID int) (*entity.VideoRating, error) {
	return s.ratingRepo.GetByUserAndVideo(ctx, userID, videoID)
}
