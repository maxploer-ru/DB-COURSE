package repository

import (
	"ZVideo/internal/domain"
	mongoinfra "ZVideo/internal/infrastructure/db/mongo"
	"ZVideo/internal/infrastructure/db/mongo/mappers"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ViewingRepository struct {
	db       *mongo.Database
	viewings *mongo.Collection
}

func NewViewingRepository(db *mongo.Database) *ViewingRepository {
	return &ViewingRepository{
		db:       db,
		viewings: db.Collection(mongoinfra.CollectionViewings),
	}
}

func (r *ViewingRepository) Create(ctx context.Context, viewing *domain.Viewing) error {
	id, err := mongoinfra.NextID(ctx, r.db, mongoinfra.CollectionViewings)
	if err != nil {
		return err
	}
	viewing.ID = id

	model := mappers.FromDomainViewing(viewing)
	if _, err := r.viewings.InsertOne(ctx, model); err != nil {
		return fmt.Errorf("create viewing: %w", err)
	}
	return nil
}

func (r *ViewingRepository) GetTotalViews(ctx context.Context, videoID int) (int, error) {
	count, err := r.viewings.CountDocuments(ctx, bson.M{"video_id": videoID})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
