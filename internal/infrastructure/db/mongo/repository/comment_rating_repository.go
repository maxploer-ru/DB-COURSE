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

type CommentRatingRepository struct {
	db      *mongo.Database
	ratings *mongo.Collection
}

func NewCommentRatingRepository(db *mongo.Database) *CommentRatingRepository {
	return &CommentRatingRepository{
		db:      db,
		ratings: db.Collection(mongoinfra.CollectionCommentRatings),
	}
}

func (r *CommentRatingRepository) Create(ctx context.Context, rating *domain.CommentRating) error {
	id := compositeID(rating.UserID, rating.CommentID)
	model := mappers.FromDomainCommentRating(rating, id)
	model.RatedAt = time.Now()

	if _, err := r.ratings.InsertOne(ctx, model); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrAlreadyRated
		}
		return fmt.Errorf("create comment rating: %w", err)
	}
	return nil
}

func (r *CommentRatingRepository) Update(ctx context.Context, rating *domain.CommentRating) error {
	id := compositeID(rating.UserID, rating.CommentID)
	update := bson.M{"$set": bson.M{"liked": rating.Liked, "rated_at": time.Now()}}
	res, err := r.ratings.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("update comment rating: %w", err)
	}
	if res.MatchedCount == 0 {
		return domain.ErrCommentRatingNotFound
	}
	return nil
}

func (r *CommentRatingRepository) Delete(ctx context.Context, userID, commentID int) error {
	id := compositeID(userID, commentID)
	res, err := r.ratings.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete comment rating: %w", err)
	}
	if res.DeletedCount == 0 {
		return domain.ErrCommentRatingNotFound
	}
	return nil
}

func (r *CommentRatingRepository) GetByUserAndComment(ctx context.Context, userID, commentID int) (*domain.CommentRating, error) {
	id := compositeID(userID, commentID)
	var model models.CommentRating
	if err := r.ratings.FindOne(ctx, bson.M{"_id": id}).Decode(&model); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return mappers.ToDomainCommentRating(&model), nil
}

func (r *CommentRatingRepository) GetStats(ctx context.Context, commentID int) (likes, dislikes int64, err error) {
	likes, err = r.ratings.CountDocuments(ctx, bson.M{"comment_id": commentID, "liked": true})
	if err != nil {
		return 0, 0, err
	}
	dislikes, err = r.ratings.CountDocuments(ctx, bson.M{"comment_id": commentID, "liked": false})
	if err != nil {
		return 0, 0, err
	}
	return likes, dislikes, nil
}

func compositeID(parts ...int) string {
	id := ""
	for i, part := range parts {
		if i > 0 {
			id += ":"
		}
		id += fmt.Sprintf("%d", part)
	}
	return id
}
