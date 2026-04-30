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
)

type ChannelRepository struct {
	db       *mongo.Database
	channels *mongo.Collection
}

func NewChannelRepository(db *mongo.Database) *ChannelRepository {
	return &ChannelRepository{db: db, channels: db.Collection(mongoinfra.CollectionChannels)}
}

func (r *ChannelRepository) Create(ctx context.Context, channel *domain.Channel) error {
	id, err := mongoinfra.NextID(ctx, r.db, mongoinfra.CollectionChannels)
	if err != nil {
		return err
	}
	channel.ID = id

	model := mappers.FromDomainChannel(channel)
	if _, err := r.channels.InsertOne(ctx, model); err != nil {
		return fmt.Errorf("create channel failed: %w", err)
	}
	return nil
}

func (r *ChannelRepository) GetByID(ctx context.Context, id int) (*domain.Channel, error) {
	var model models.Channel
	if err := r.channels.FindOne(ctx, bson.M{"_id": id}).Decode(&model); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("get channel failed: %w", err)
	}
	return mappers.ToDomainChannel(&model), nil
}

func (r *ChannelRepository) GetByUserID(ctx context.Context, userID int) (*domain.Channel, error) {
	var model models.Channel
	if err := r.channels.FindOne(ctx, bson.M{"user_id": userID}).Decode(&model); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("get channel by user failed: %w", err)
	}
	return mappers.ToDomainChannel(&model), nil
}

func (r *ChannelRepository) GetByName(ctx context.Context, name string) (*domain.Channel, error) {
	var model models.Channel
	if err := r.channels.FindOne(ctx, bson.M{"name": name}).Decode(&model); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("get channel by name failed: %w", err)
	}
	return mappers.ToDomainChannel(&model), nil
}

func (r *ChannelRepository) Update(ctx context.Context, channel *domain.Channel) error {
	update := bson.M{"$set": bson.M{"name": channel.Name, "description": channel.Description}}
	res, err := r.channels.UpdateOne(ctx, bson.M{"_id": channel.ID}, update)
	if err != nil {
		return fmt.Errorf("update channel failed: %w", err)
	}
	if res.MatchedCount == 0 {
		return domain.ErrChannelNotFound
	}
	return nil
}

func (r *ChannelRepository) Delete(ctx context.Context, id int) error {
	res, err := r.channels.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete channel failed: %w", err)
	}
	if res.DeletedCount == 0 {
		return domain.ErrChannelNotFound
	}
	return nil
}

func (r *ChannelRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	count, err := r.channels.CountDocuments(ctx, bson.M{"name": name})
	return count > 0, err
}
