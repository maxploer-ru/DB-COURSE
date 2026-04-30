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

type CommunityRepository struct {
	db       *mongo.Database
	posts    *mongo.Collection
	comments *mongo.Collection
}

func NewCommunityRepository(db *mongo.Database) *CommunityRepository {
	return &CommunityRepository{
		db:       db,
		posts:    db.Collection(mongoinfra.CollectionCommunityPosts),
		comments: db.Collection(mongoinfra.CollectionCommunityComments),
	}
}

func (r *CommunityRepository) CreatePost(ctx context.Context, post *domain.CommunityPost) error {
	id, err := mongoinfra.NextID(ctx, r.db, mongoinfra.CollectionCommunityPosts)
	if err != nil {
		return err
	}
	post.ID = id

	model := mappers.FromDomainCommunityPost(post)
	if _, err := r.posts.InsertOne(ctx, model); err != nil {
		return fmt.Errorf("create community post: %w", err)
	}
	return nil
}

func (r *CommunityRepository) GetPostByID(ctx context.Context, id int) (*domain.CommunityPost, error) {
	var model models.CommunityPost
	if err := r.posts.FindOne(ctx, bson.M{"_id": id}).Decode(&model); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("get community post by id: %w", err)
	}
	return mappers.ToDomainCommunityPost(&model), nil
}

func (r *CommunityRepository) ListPostsByChannel(ctx context.Context, channelID int, limit, offset int) ([]*domain.CommunityPost, error) {
	filter := bson.M{"channel_id": channelID}
	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset)).SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.posts.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("list community posts by channel: %w", err)
	}
	defer cursor.Close(ctx)

	modelsList := make([]models.CommunityPost, 0)
	for cursor.Next(ctx) {
		var model models.CommunityPost
		if err := cursor.Decode(&model); err != nil {
			return nil, fmt.Errorf("decode community post: %w", err)
		}
		modelsList = append(modelsList, model)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("list community posts cursor: %w", err)
	}

	result := make([]*domain.CommunityPost, 0, len(modelsList))
	for _, model := range modelsList {
		result = append(result, mappers.ToDomainCommunityPost(&model))
	}
	return result, nil
}

func (r *CommunityRepository) UpdatePost(ctx context.Context, post *domain.CommunityPost) error {
	res, err := r.posts.UpdateOne(ctx, bson.M{"_id": post.ID}, bson.M{"$set": bson.M{"content": post.Content}})
	if err != nil {
		return fmt.Errorf("update community post: %w", err)
	}
	if res.MatchedCount == 0 {
		return domain.ErrCommunityPostNotFound
	}
	return nil
}

func (r *CommunityRepository) DeletePost(ctx context.Context, id int) error {
	res, err := r.posts.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete community post: %w", err)
	}
	if res.DeletedCount == 0 {
		return domain.ErrCommunityPostNotFound
	}
	return nil
}

func (r *CommunityRepository) CreateComment(ctx context.Context, comment *domain.CommunityComment) error {
	id, err := mongoinfra.NextID(ctx, r.db, mongoinfra.CollectionCommunityComments)
	if err != nil {
		return err
	}
	comment.ID = id

	model := mappers.FromDomainCommunityComment(comment)
	if _, err := r.comments.InsertOne(ctx, model); err != nil {
		return fmt.Errorf("create community comment: %w", err)
	}
	return nil
}

func (r *CommunityRepository) GetCommentByID(ctx context.Context, id int) (*domain.CommunityComment, error) {
	var model models.CommunityComment
	if err := r.comments.FindOne(ctx, bson.M{"_id": id}).Decode(&model); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("get community comment by id: %w", err)
	}
	return mappers.ToDomainCommunityComment(&model), nil
}

func (r *CommunityRepository) ListCommentsByPost(ctx context.Context, postID int, limit, offset int) ([]*domain.CommunityComment, error) {
	filter := bson.M{"post_id": postID}
	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset)).SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.comments.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("list community comments by post: %w", err)
	}
	defer cursor.Close(ctx)

	modelsList := make([]models.CommunityComment, 0)
	for cursor.Next(ctx) {
		var model models.CommunityComment
		if err := cursor.Decode(&model); err != nil {
			return nil, fmt.Errorf("decode community comment: %w", err)
		}
		modelsList = append(modelsList, model)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("list community comments cursor: %w", err)
	}

	result := make([]*domain.CommunityComment, 0, len(modelsList))
	for _, model := range modelsList {
		result = append(result, mappers.ToDomainCommunityComment(&model))
	}
	return result, nil
}

func (r *CommunityRepository) UpdateComment(ctx context.Context, comment *domain.CommunityComment) error {
	res, err := r.comments.UpdateOne(ctx, bson.M{"_id": comment.ID}, bson.M{"$set": bson.M{"content": comment.Content}})
	if err != nil {
		return fmt.Errorf("update community comment: %w", err)
	}
	if res.MatchedCount == 0 {
		return domain.ErrCommunityCommentNotFound
	}
	return nil
}

func (r *CommunityRepository) DeleteComment(ctx context.Context, id int) error {
	res, err := r.comments.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete community comment: %w", err)
	}
	if res.DeletedCount == 0 {
		return domain.ErrCommunityCommentNotFound
	}
	return nil
}
