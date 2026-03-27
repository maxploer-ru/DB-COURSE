package service

import (
	"ZVideo/internal/domain/video/entity"
	"ZVideo/internal/domain/video/repository"
	"context"
)

type VideoStatsService interface {
	GetVideoStats(ctx context.Context, videoID int) (*entity.VideoStats, error)
}

type videoStatsService struct {
	repo  repository.VideoStatsRepository
	cache repository.VideoStatsCache
}

func NewVideoStatsService(
	repo repository.VideoStatsRepository,
	cache repository.VideoStatsCache,
) VideoStatsService {
	return &videoStatsService{
		repo:  repo,
		cache: cache,
	}
}

func (s *videoStatsService) GetVideoStats(ctx context.Context, videoID int) (*entity.VideoStats, error) {

	if stats, found, err := s.cache.Get(ctx, videoID); err == nil && found {
		return stats, nil
	}

	views, err := s.repo.GetViewsCount(ctx, videoID)
	if err != nil {
		return nil, err
	}

	likes, dislikes, err := s.repo.GetLikesDislikes(ctx, videoID)
	if err != nil {
		return nil, err
	}

	comments, err := s.repo.GetCommentsCount(ctx, videoID)
	if err != nil {
		return nil, err
	}

	stats := &entity.VideoStats{
		VideoID:       videoID,
		ViewsCount:    views,
		LikesCount:    likes,
		DislikesCount: dislikes,
		CommentsCount: comments,
	}

	_ = s.cache.Set(ctx, videoID, stats)
	return stats, nil
}
