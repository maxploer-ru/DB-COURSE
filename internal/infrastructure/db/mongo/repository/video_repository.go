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

type VideoRepository struct {
	db       *mongo.Database
	videos   *mongo.Collection
	channels *mongo.Collection
}

func NewVideoRepository(db *mongo.Database) *VideoRepository {
	return &VideoRepository{
		db:       db,
		videos:   db.Collection(mongoinfra.CollectionVideos),
		channels: db.Collection(mongoinfra.CollectionChannels),
	}
}

func (r *VideoRepository) Create(ctx context.Context, video *domain.Video) error {
	id, err := mongoinfra.NextID(ctx, r.db, mongoinfra.CollectionVideos)
	if err != nil {
		return err
	}
	video.ID = id

	model := mappers.FromDomainVideo(video)
	if _, err := r.videos.InsertOne(ctx, model); err != nil {
		return fmt.Errorf("create video: %w", err)
	}
	return nil
}

func (r *VideoRepository) GetByID(ctx context.Context, id int) (*domain.Video, error) {
	var model models.Video
	if err := r.videos.FindOne(ctx, bson.M{"_id": id}).Decode(&model); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("get video by id: %w", err)
	}

	channelName, err := r.resolveChannelName(ctx, model.ChannelID)
	if err != nil {
		return nil, err
	}

	return mappers.ToDomainVideo(&model, channelName), nil
}

func (r *VideoRepository) Update(ctx context.Context, video *domain.Video) error {
	update := bson.M{
		"title":       video.Title,
		"description": video.Description,
		"filepath":    video.Filepath,
	}
	res, err := r.videos.UpdateOne(ctx, bson.M{"_id": video.ID}, bson.M{"$set": update})
	if err != nil {
		return fmt.Errorf("update video: %w", err)
	}
	if res.MatchedCount == 0 {
		return domain.ErrVideoNotFound
	}
	return nil
}

func (r *VideoRepository) Delete(ctx context.Context, id int) error {
	res, err := r.videos.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete video: %w", err)
	}
	if res.DeletedCount == 0 {
		return domain.ErrVideoNotFound
	}
	return nil
}

func (r *VideoRepository) List(ctx context.Context, limit, offset int, sort domain.VideoSort) ([]*domain.Video, error) {
	return r.listVideos(ctx, bson.M{}, limit, offset, sort)
}

func (r *VideoRepository) ListByChannel(ctx context.Context, channelID int, limit, offset int, sort domain.VideoSort) ([]*domain.Video, error) {
	filter := bson.M{"channel_id": channelID}
	return r.listVideos(ctx, filter, limit, offset, sort)
}

func (r *VideoRepository) listVideos(ctx context.Context, filter bson.M, limit, offset int, sort domain.VideoSort) ([]*domain.Video, error) {
	if sort == domain.VideoSortNewest {
		opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset)).SetSort(bson.D{{Key: "created_at", Value: -1}})
		cursor, err := r.videos.Find(ctx, filter, opts)
		if err != nil {
			return nil, fmt.Errorf("list videos: %w", err)
		}
		defer cursor.Close(ctx)

		modelsList := make([]models.Video, 0)
		channelIDs := make([]int, 0)
		for cursor.Next(ctx) {
			var model models.Video
			if err := cursor.Decode(&model); err != nil {
				return nil, fmt.Errorf("decode video: %w", err)
			}
			modelsList = append(modelsList, model)
			channelIDs = append(channelIDs, model.ChannelID)
		}
		if err := cursor.Err(); err != nil {
			return nil, fmt.Errorf("list videos cursor: %w", err)
		}

		channelNames, err := loadChannelNames(ctx, r.channels, channelIDs)
		if err != nil {
			return nil, err
		}

		result := make([]*domain.Video, 0, len(modelsList))
		for _, model := range modelsList {
			result = append(result, mappers.ToDomainVideo(&model, channelNames[model.ChannelID]))
		}
		return result, nil
	}

	pipeline := buildVideoSortPipeline(filter, sort, limit, offset)
	cursor, err := r.videos.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("aggregate videos: %w", err)
	}
	defer cursor.Close(ctx)

	modelsList := make([]models.Video, 0)
	channelIDs := make([]int, 0)
	for cursor.Next(ctx) {
		var model models.Video
		if err := cursor.Decode(&model); err != nil {
			return nil, fmt.Errorf("decode video: %w", err)
		}
		modelsList = append(modelsList, model)
		channelIDs = append(channelIDs, model.ChannelID)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("list videos cursor: %w", err)
	}

	channelNames, err := loadChannelNames(ctx, r.channels, channelIDs)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Video, 0, len(modelsList))
	for _, model := range modelsList {
		result = append(result, mappers.ToDomainVideo(&model, channelNames[model.ChannelID]))
	}
	return result, nil
}

func buildVideoSortPipeline(filter bson.M, sort domain.VideoSort, limit, offset int) mongo.Pipeline {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         mongoinfra.CollectionViewings,
			"localField":   "_id",
			"foreignField": "video_id",
			"as":           "viewings",
		}}},
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         mongoinfra.CollectionVideoRatings,
			"localField":   "_id",
			"foreignField": "video_id",
			"as":           "ratings",
		}}},
		bson.D{{Key: "$addFields", Value: bson.M{
			"views_count": bson.M{"$size": "$viewings"},
			"rating_score": bson.M{
				"$subtract": bson.A{
					bson.M{"$size": bson.M{
						"$filter": bson.M{
							"input": "$ratings",
							"as":    "r",
							"cond":  bson.M{"$eq": bson.A{"$$r.liked", true}},
						},
					}},
					bson.M{"$size": bson.M{
						"$filter": bson.M{
							"input": "$ratings",
							"as":    "r",
							"cond":  bson.M{"$eq": bson.A{"$$r.liked", false}},
						},
					},
					},
				},
			},
		}}},
	}

	sortField := "created_at"
	switch sort {
	case domain.VideoSortViews:
		sortField = "views_count"
	case domain.VideoSortRating:
		sortField = "rating_score"
	}

	pipeline = append(pipeline,
		bson.D{{Key: "$sort", Value: bson.D{{Key: sortField, Value: -1}, {Key: "created_at", Value: -1}}}},
		bson.D{{Key: "$skip", Value: int64(offset)}},
		bson.D{{Key: "$limit", Value: int64(limit)}},
	)
	return pipeline
}

func (r *VideoRepository) ListFilepathsByChannel(ctx context.Context, channelID int) ([]string, error) {
	opts := options.Find().SetProjection(bson.M{"filepath": 1})
	cursor, err := r.videos.Find(ctx, bson.M{"channel_id": channelID}, opts)
	if err != nil {
		return nil, fmt.Errorf("list filepaths by channel: %w", err)
	}
	defer cursor.Close(ctx)

	filepaths := make([]string, 0)
	for cursor.Next(ctx) {
		var model struct {
			Filepath string `bson:"filepath"`
		}
		if err := cursor.Decode(&model); err != nil {
			return nil, fmt.Errorf("decode filepath: %w", err)
		}
		filepaths = append(filepaths, model.Filepath)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("list filepaths cursor: %w", err)
	}
	return filepaths, nil
}

func (r *VideoRepository) resolveChannelName(ctx context.Context, channelID int) (string, error) {
	var channel struct {
		Name string `bson:"name"`
	}
	if err := r.channels.FindOne(ctx, bson.M{"_id": channelID}, options.FindOne().SetProjection(bson.M{"name": 1})).Decode(&channel); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", nil
		}
		return "", fmt.Errorf("get channel name: %w", err)
	}
	return channel.Name, nil
}
