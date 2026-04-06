package repositories

import (
	"context"

	"gorm.io/gorm"
)

type VideoStatsRepository struct {
	db *gorm.DB
}

func NewVideoStatsRepository(db *gorm.DB) *VideoStatsRepository {
	return &VideoStatsRepository{
		db: db,
	}
}

func (r *VideoStatsRepository) GetViewsCount(ctx context.Context, videoID int) (int, error) {
	var count int64

	err := r.db.WithContext(ctx).Table("viewings").Where("video_id = ?", videoID).Count(&count).Error
	return int(count), err
}

func (r *VideoStatsRepository) GetLikesDislikes(ctx context.Context, videoID int) (int, int, error) {
	type result struct {
		IsLike bool
		Count  int
	}
	var results []result

	err := r.db.WithContext(ctx).Table("video_ratings").
		Select("is_like, COUNT(*) as count").
		Where("video_id = ?", videoID).
		Group("is_like").
		Scan(&results).Error

	if err != nil {
		return 0, 0, err
	}

	var likes, dislikes int
	for _, res := range results {
		if res.IsLike {
			likes = res.Count
		} else {
			dislikes = res.Count
		}
	}

	return likes, dislikes, nil
}

func (r *VideoStatsRepository) GetCommentsCount(ctx context.Context, videoID int) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("comments").Where("video_id = ?", videoID).Count(&count).Error
	return int(count), err
}
