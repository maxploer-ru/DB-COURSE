package repository

import (
	"ZVideo/internal/domain"
	mongoinfra "ZVideo/internal/infrastructure/db/mongo"
	"ZVideo/internal/infrastructure/db/mongo/mappers"
	"ZVideo/internal/infrastructure/db/mongo/models"
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CommentRepository struct {
	db       *mongo.Database
	comments *mongo.Collection
	users    *mongo.Collection
}

func NewCommentRepository(db *mongo.Database) *CommentRepository {
	return &CommentRepository{
		db:       db,
		comments: db.Collection(mongoinfra.CollectionComments),
		users:    db.Collection(mongoinfra.CollectionUsers),
	}
}

func (r *CommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	id, err := mongoinfra.NextID(ctx, r.db, mongoinfra.CollectionComments)
	if err != nil {
		return err
	}
	comment.ID = id

	model := mappers.FromDomainComment(comment)
	if _, err := r.comments.InsertOne(ctx, model); err != nil {
		return fmt.Errorf("create comment: %w", err)
	}
	return nil
}

func (r *CommentRepository) GetByID(ctx context.Context, id int) (*domain.Comment, error) {
	var model models.Comment
	if err := r.comments.FindOne(ctx, bson.M{"_id": id}).Decode(&model); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("get comment by id: %w", err)
	}

	username, err := r.resolveUsername(ctx, model.UserID)
	if err != nil {
		return nil, err
	}
	return mappers.ToDomainComment(&model, username), nil
}

func (r *CommentRepository) ListByVideo(ctx context.Context, videoID int, limit, offset int) ([]*domain.Comment, error) {
	filter := bson.M{"video_id": videoID}
	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset)).SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.comments.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("list comments by video: %w", err)
	}
	defer cursor.Close(ctx)

	modelsList := make([]models.Comment, 0)
	userIDs := make([]int, 0)
	for cursor.Next(ctx) {
		var model models.Comment
		if err := cursor.Decode(&model); err != nil {
			return nil, fmt.Errorf("decode comment: %w", err)
		}
		modelsList = append(modelsList, model)
		userIDs = append(userIDs, model.UserID)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("list comments cursor: %w", err)
	}

	usernames, err := loadUsernames(ctx, r.users, userIDs)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Comment, 0, len(modelsList))
	for _, model := range modelsList {
		result = append(result, mappers.ToDomainComment(&model, usernames[model.UserID]))
	}
	return result, nil
}

func (r *CommentRepository) Update(ctx context.Context, comment *domain.Comment) error {
	res, err := r.comments.UpdateOne(ctx, bson.M{"_id": comment.ID}, bson.M{"$set": bson.M{"content": comment.Content}})
	if err != nil {
		return fmt.Errorf("update comment: %w", err)
	}
	if res.MatchedCount == 0 {
		return domain.ErrCommentNotFound
	}
	return nil
}

func (r *CommentRepository) Delete(ctx context.Context, id int) error {
	res, err := r.comments.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	if res.DeletedCount == 0 {
		return domain.ErrCommentNotFound
	}
	return nil
}

func (r *CommentRepository) CountByVideo(ctx context.Context, videoID int) (int64, error) {
	return r.comments.CountDocuments(ctx, bson.M{"video_id": videoID})
}

func (r *CommentRepository) resolveUsername(ctx context.Context, userID int) (string, error) {
	var user struct {
		Username string `bson:"username"`
	}
	if err := r.users.FindOne(ctx, bson.M{"_id": userID}, options.FindOne().SetProjection(bson.M{"username": 1})).Decode(&user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", nil
		}
		return "", fmt.Errorf("get username: %w", err)
	}
	return user.Username, nil
}
