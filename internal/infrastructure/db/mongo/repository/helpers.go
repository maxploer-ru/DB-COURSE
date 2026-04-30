package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func duplicateIndexName(err error) string {
	var we mongo.WriteException
	if errors.As(err, &we) {
		for _, e := range we.WriteErrors {
			if e.Code == 11000 || e.Code == 11001 {
				msg := e.Message
				if idx := strings.Index(msg, "index: "); idx != -1 {
					rest := msg[idx+len("index: "):]
					if end := strings.Index(rest, " "); end != -1 {
						return rest[:end]
					}
					return rest
				}
			}
		}
	}
	return ""
}

func loadUsernames(ctx context.Context, usersColl *mongo.Collection, userIDs []int) (map[int]string, error) {
	if len(userIDs) == 0 {
		return map[int]string{}, nil
	}
	filter := bson.M{"_id": bson.M{"$in": userIDs}}
	opts := options.Find().SetProjection(bson.M{"username": 1})
	cursor, err := usersColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("load usernames: %w", err)
	}
	defer cursor.Close(ctx)

	result := make(map[int]string, len(userIDs))
	for cursor.Next(ctx) {
		var u struct {
			ID       int    `bson:"_id"`
			Username string `bson:"username"`
		}
		if err := cursor.Decode(&u); err != nil {
			return nil, fmt.Errorf("decode usernames: %w", err)
		}
		result[u.ID] = u.Username
	}
	return result, cursor.Err()
}

func loadChannelNames(ctx context.Context, channelsColl *mongo.Collection, channelIDs []int) (map[int]string, error) {
	if len(channelIDs) == 0 {
		return map[int]string{}, nil
	}
	filter := bson.M{"_id": bson.M{"$in": channelIDs}}
	opts := options.Find().SetProjection(bson.M{"name": 1})
	cursor, err := channelsColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("load channel names: %w", err)
	}
	defer cursor.Close(ctx)

	result := make(map[int]string, len(channelIDs))
	for cursor.Next(ctx) {
		var ch struct {
			ID   int    `bson:"_id"`
			Name string `bson:"name"`
		}
		if err := cursor.Decode(&ch); err != nil {
			return nil, fmt.Errorf("decode channel names: %w", err)
		}
		result[ch.ID] = ch.Name
	}
	return result, cursor.Err()
}

func loadVideoTitles(ctx context.Context, videosColl *mongo.Collection, videoIDs []int) (map[int]string, error) {
	if len(videoIDs) == 0 {
		return map[int]string{}, nil
	}
	filter := bson.M{"_id": bson.M{"$in": videoIDs}}
	opts := options.Find().SetProjection(bson.M{"title": 1})
	cursor, err := videosColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("load video titles: %w", err)
	}
	defer cursor.Close(ctx)

	result := make(map[int]string, len(videoIDs))
	for cursor.Next(ctx) {
		var v struct {
			ID    int    `bson:"_id"`
			Title string `bson:"title"`
		}
		if err := cursor.Decode(&v); err != nil {
			return nil, fmt.Errorf("decode video titles: %w", err)
		}
		result[v.ID] = v.Title
	}
	return result, cursor.Err()
}
