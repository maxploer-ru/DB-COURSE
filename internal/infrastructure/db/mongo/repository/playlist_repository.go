package repository

import (
	"ZVideo/internal/domain"
	mongoinfra "ZVideo/internal/infrastructure/db/mongo"
	"ZVideo/internal/infrastructure/db/mongo/mappers"
	"ZVideo/internal/infrastructure/db/mongo/models"
	"context"
	"errors"
	"fmt"
	"sort"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlaylistRepository struct {
	db        *mongo.Database
	playlists *mongo.Collection
	videos    *mongo.Collection
}

func NewPlaylistRepository(db *mongo.Database) *PlaylistRepository {
	return &PlaylistRepository{
		db:        db,
		playlists: db.Collection(mongoinfra.CollectionPlaylists),
		videos:    db.Collection(mongoinfra.CollectionVideos),
	}
}

func (r *PlaylistRepository) Create(ctx context.Context, playlist *domain.Playlist) error {
	id, err := mongoinfra.NextID(ctx, r.db, mongoinfra.CollectionPlaylists)
	if err != nil {
		return err
	}
	playlist.ID = id

	model := mappers.FromDomainPlaylist(playlist, nil)
	if _, err := r.playlists.InsertOne(ctx, model); err != nil {
		return fmt.Errorf("create playlist failed: %w", err)
	}
	return nil
}

func (r *PlaylistRepository) GetByID(ctx context.Context, playlistID int) (*domain.Playlist, error) {
	var model models.Playlist
	if err := r.playlists.FindOne(ctx, bson.M{"_id": playlistID}).Decode(&model); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("get playlist failed: %w", err)
	}

	items, err := r.toDomainItems(ctx, model.Items)
	if err != nil {
		return nil, err
	}
	return mappers.ToDomainPlaylist(&model, items), nil
}

func (r *PlaylistRepository) ListByChannel(ctx context.Context, channelID int, limit, offset int) ([]*domain.Playlist, error) {
	filter := bson.M{"channel_id": channelID}
	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset)).SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.playlists.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("list playlists by channel failed: %w", err)
	}
	defer cursor.Close(ctx)

	modelsList := make([]models.Playlist, 0)
	for cursor.Next(ctx) {
		var model models.Playlist
		if err := cursor.Decode(&model); err != nil {
			return nil, fmt.Errorf("decode playlist: %w", err)
		}
		modelsList = append(modelsList, model)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("list playlists cursor: %w", err)
	}

	result := make([]*domain.Playlist, 0, len(modelsList))
	for _, model := range modelsList {
		items, err := r.toDomainItems(ctx, model.Items)
		if err != nil {
			return nil, err
		}
		result = append(result, mappers.ToDomainPlaylist(&model, items))
	}
	return result, nil
}

func (r *PlaylistRepository) Update(ctx context.Context, playlist *domain.Playlist) error {
	update := bson.M{"$set": bson.M{"name": playlist.Name, "description": playlist.Description}}
	res, err := r.playlists.UpdateOne(ctx, bson.M{"_id": playlist.ID}, update)
	if err != nil {
		return fmt.Errorf("update playlist failed: %w", err)
	}
	if res.MatchedCount == 0 {
		return domain.ErrPlaylistNotFound
	}
	return nil
}

func (r *PlaylistRepository) Delete(ctx context.Context, playlistID int) error {
	res, err := r.playlists.DeleteOne(ctx, bson.M{"_id": playlistID})
	if err != nil {
		return fmt.Errorf("delete playlist failed: %w", err)
	}
	if res.DeletedCount == 0 {
		return domain.ErrPlaylistNotFound
	}
	return nil
}

func (r *PlaylistRepository) AddVideo(ctx context.Context, playlistID, videoID int) error {
	var playlist models.Playlist
	if err := r.playlists.FindOne(ctx, bson.M{"_id": playlistID}).Decode(&playlist); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.ErrPlaylistNotFound
		}
		return fmt.Errorf("get playlist for add video failed: %w", err)
	}

	maxNumber := 0
	for _, item := range playlist.Items {
		if item.VideoID == videoID {
			return nil
		}
		if item.Number > maxNumber {
			maxNumber = item.Number
		}
	}

	newItem := models.PlaylistItem{VideoID: videoID, Number: maxNumber + 1}
	update := bson.M{"$push": bson.M{"items": newItem}}
	_, err := r.playlists.UpdateOne(ctx, bson.M{"_id": playlistID}, update)
	if err != nil {
		return fmt.Errorf("add video to playlist failed: %w", err)
	}
	return nil
}

func (r *PlaylistRepository) RemoveVideo(ctx context.Context, playlistID, videoID int) error {
	update := bson.M{"$pull": bson.M{"items": bson.M{"video_id": videoID}}}
	if _, err := r.playlists.UpdateOne(ctx, bson.M{"_id": playlistID}, update); err != nil {
		return fmt.Errorf("remove video from playlist failed: %w", err)
	}
	return nil
}

func (r *PlaylistRepository) toDomainItems(ctx context.Context, items []models.PlaylistItem) ([]domain.PlaylistItem, error) {
	if len(items) == 0 {
		return nil, nil
	}
	videoIDs := make([]int, 0, len(items))
	for _, item := range items {
		videoIDs = append(videoIDs, item.VideoID)
	}
	videoTitles, err := loadVideoTitles(ctx, r.videos, videoIDs)
	if err != nil {
		return nil, err
	}

	result := make([]domain.PlaylistItem, 0, len(items))
	for _, item := range items {
		result = append(result, domain.PlaylistItem{
			VideoID:    item.VideoID,
			VideoTitle: videoTitles[item.VideoID],
			Number:     item.Number,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Number < result[j].Number
	})
	return result, nil
}
