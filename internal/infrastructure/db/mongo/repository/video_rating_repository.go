package repository

import (
	"ZVideo/internal/domain"
	mongoinfra "ZVideo/internal/infrastructure/db/mongo"
	"ZVideo/internal/infrastructure/db/mongo/mappers"
	"ZVideo/internal/infrastructure/db/mongo/models"
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type VideoRatingRepository struct {
	db      *mongo.Database
	ratings *mongo.Collection
}

func NewVideoRatingRepository(db *mongo.Database) *VideoRatingRepository {
	return &VideoRatingRepository{
		db:      db,
		ratings: db.Collection(mongoinfra.CollectionVideoRatings),
	}
}

func (r *VideoRatingRepository) Create(ctx context.Context, rating *domain.VideoRating) error {
	id := compositeID(rating.UserID, rating.VideoID)
	model := mappers.FromDomainVideoRating(rating, id)
	model.RatedAt = time.Now()

	if _, err := r.ratings.InsertOne(ctx, model); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrAlreadyRated
		}
		return fmt.Errorf("create video rating: %w", err)
	}
	return nil
}

func (r *VideoRatingRepository) Update(ctx context.Context, rating *domain.VideoRating) error {
	id := compositeID(rating.UserID, rating.VideoID)
	update := bson.M{"$set": bson.M{"liked": rating.Liked, "rated_at": time.Now()}}
	res, err := r.ratings.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("update video rating: %w", err)
	}
	if res.MatchedCount == 0 {
		return domain.ErrRatingNotFound
	}
	return nil
}

func (r *VideoRatingRepository) Delete(ctx context.Context, userID, videoID int) error {
	id := compositeID(userID, videoID)
	res, err := r.ratings.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete video rating: %w", err)
	}
	if res.DeletedCount == 0 {
		return domain.ErrRatingNotFound
	}
	return nil
}

func (r *VideoRatingRepository) GetByUserAndVideo(ctx context.Context, userID, videoID int) (*domain.VideoRating, error) {
	id := compositeID(userID, videoID)
	var model models.VideoRating
	if err := r.ratings.FindOne(ctx, bson.M{"_id": id}).Decode(&model); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return mappers.ToDomainVideoRating(&model), nil
}

func (r *VideoRatingRepository) GetStats(ctx context.Context, videoID int) (likes, dislikes int, err error) {
	likesCount, err := r.ratings.CountDocuments(ctx, bson.M{"video_id": videoID, "liked": true})
	if err != nil {
		return 0, 0, err
	}
	dislikesCount, err := r.ratings.CountDocuments(ctx, bson.M{"video_id": videoID, "liked": false})
	if err != nil {
		return 0, 0, err
	}
	return int(likesCount), int(dislikesCount), nil
}
