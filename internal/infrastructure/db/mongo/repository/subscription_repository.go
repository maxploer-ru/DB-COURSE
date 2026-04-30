package repository

import (
	"ZVideo/internal/domain"
	mongoinfra "ZVideo/internal/infrastructure/db/mongo"
	"ZVideo/internal/infrastructure/db/mongo/mappers"
	"ZVideo/internal/infrastructure/db/mongo/models"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SubscriptionRepository struct {
	subs *mongo.Collection
}

func NewSubscriptionRepository(db *mongo.Database) *SubscriptionRepository {
	return &SubscriptionRepository{subs: db.Collection(mongoinfra.CollectionSubscriptions)}
}

func (r *SubscriptionRepository) Subscribe(ctx context.Context, userID, channelID int) (bool, error) {
	id := compositeID(userID, channelID)
	model := &models.Subscription{
		ID:             id,
		UserID:         userID,
		ChannelID:      channelID,
		NewVideosCount: 0,
		SubscribedAt:   time.Now(),
	}
	if _, err := r.subs.InsertOne(ctx, model); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return false, nil
		}
		return false, fmt.Errorf("subscribe failed: %w", err)
	}
	return true, nil
}

func (r *SubscriptionRepository) Unsubscribe(ctx context.Context, userID, channelID int) (bool, error) {
	id := compositeID(userID, channelID)
	res, err := r.subs.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return false, fmt.Errorf("unsubscribe failed: %w", err)
	}
	return res.DeletedCount > 0, nil
}

func (r *SubscriptionRepository) IsSubscribed(ctx context.Context, userID, channelID int) (bool, error) {
	id := compositeID(userID, channelID)
	count, err := r.subs.CountDocuments(ctx, bson.M{"_id": id})
	return count > 0, err
}

func (r *SubscriptionRepository) GetSubscribersCount(ctx context.Context, channelID int) (int, error) {
	count, err := r.subs.CountDocuments(ctx, bson.M{"channel_id": channelID})
	return int(count), err
}

func (r *SubscriptionRepository) GetUserSubscriptions(ctx context.Context, userID int, limit, offset int) ([]*domain.Subscription, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset)).SetSort(bson.D{{Key: "subscribed_at", Value: -1}})
	cursor, err := r.subs.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	modelsList := make([]models.Subscription, 0)
	for cursor.Next(ctx) {
		var model models.Subscription
		if err := cursor.Decode(&model); err != nil {
			return nil, err
		}
		modelsList = append(modelsList, model)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	result := make([]*domain.Subscription, 0, len(modelsList))
	for _, model := range modelsList {
		result = append(result, mappers.ToDomainSubscription(&model))
	}
	return result, nil
}

func (r *SubscriptionRepository) NotifySubscribersAboutNewVideo(ctx context.Context, channelID int) error {
	update := bson.M{"$inc": bson.M{"new_videos_count": 1}}
	if _, err := r.subs.UpdateMany(ctx, bson.M{"channel_id": channelID}, update); err != nil {
		return fmt.Errorf("notify subscribers failed: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) ResetNewVideosCount(ctx context.Context, userID, channelID int) error {
	id := compositeID(userID, channelID)
	update := bson.M{"$set": bson.M{"new_videos_count": 0}}
	if _, err := r.subs.UpdateOne(ctx, bson.M{"_id": id}, update); err != nil {
		return fmt.Errorf("reset new videos count failed: %w", err)
	}
	return nil
}
