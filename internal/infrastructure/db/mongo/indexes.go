package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	CollectionUsers             = "users"
	CollectionRoles             = "roles"
	CollectionChannels          = "channels"
	CollectionVideos            = "videos"
	CollectionComments          = "comments"
	CollectionCommentRatings    = "comment_ratings"
	CollectionVideoRatings      = "video_ratings"
	CollectionSubscriptions     = "subscriptions"
	CollectionPlaylists         = "playlists"
	CollectionCommunityPosts    = "community_posts"
	CollectionCommunityComments = "community_comments"
	CollectionViewings          = "viewings"
	CollectionCounters          = "counters"
)

func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	if err := ensureUserIndexes(ctx, db.Collection(CollectionUsers)); err != nil {
		return err
	}
	if err := ensureRoleIndexes(ctx, db.Collection(CollectionRoles)); err != nil {
		return err
	}
	if err := ensureChannelIndexes(ctx, db.Collection(CollectionChannels)); err != nil {
		return err
	}
	if err := ensureVideoIndexes(ctx, db.Collection(CollectionVideos)); err != nil {
		return err
	}
	if err := ensureCommentIndexes(ctx, db.Collection(CollectionComments)); err != nil {
		return err
	}
	if err := ensureCommentRatingIndexes(ctx, db.Collection(CollectionCommentRatings)); err != nil {
		return err
	}
	if err := ensureVideoRatingIndexes(ctx, db.Collection(CollectionVideoRatings)); err != nil {
		return err
	}
	if err := ensureSubscriptionIndexes(ctx, db.Collection(CollectionSubscriptions)); err != nil {
		return err
	}
	if err := ensurePlaylistIndexes(ctx, db.Collection(CollectionPlaylists)); err != nil {
		return err
	}
	if err := ensureCommunityIndexes(ctx, db.Collection(CollectionCommunityPosts), db.Collection(CollectionCommunityComments)); err != nil {
		return err
	}
	if err := ensureViewingIndexes(ctx, db.Collection(CollectionViewings)); err != nil {
		return err
	}
	return nil
}

func ensureUserIndexes(ctx context.Context, col *mongo.Collection) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "username", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "role_id", Value: 1}}},
	}
	_, err := col.Indexes().CreateMany(ctx, indexes)
	return err
}

func ensureRoleIndexes(ctx context.Context, col *mongo.Collection) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "name", Value: 1}}, Options: options.Index().SetUnique(true)},
		{
			Keys: bson.D{{Key: "is_default", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetPartialFilterExpression(bson.M{"is_default": true}),
		},
	}
	_, err := col.Indexes().CreateMany(ctx, indexes)
	return err
}

func ensureChannelIndexes(ctx context.Context, col *mongo.Collection) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "name", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
	}
	_, err := col.Indexes().CreateMany(ctx, indexes)
	return err
}

func ensureVideoIndexes(ctx context.Context, col *mongo.Collection) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "channel_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	}
	_, err := col.Indexes().CreateMany(ctx, indexes)
	return err
}

func ensureCommentIndexes(ctx context.Context, col *mongo.Collection) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "video_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
	}
	_, err := col.Indexes().CreateMany(ctx, indexes)
	return err
}

func ensureCommentRatingIndexes(ctx context.Context, col *mongo.Collection) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "comment_id", Value: 1}}},
		{Keys: bson.D{{Key: "comment_id", Value: 1}, {Key: "liked", Value: 1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
	}
	_, err := col.Indexes().CreateMany(ctx, indexes)
	return err
}

func ensureVideoRatingIndexes(ctx context.Context, col *mongo.Collection) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "video_id", Value: 1}}},
		{Keys: bson.D{{Key: "video_id", Value: 1}, {Key: "liked", Value: 1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
	}
	_, err := col.Indexes().CreateMany(ctx, indexes)
	return err
}

func ensureSubscriptionIndexes(ctx context.Context, col *mongo.Collection) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "channel_id", Value: 1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "subscribed_at", Value: -1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
	}
	_, err := col.Indexes().CreateMany(ctx, indexes)
	return err
}

func ensurePlaylistIndexes(ctx context.Context, col *mongo.Collection) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "channel_id", Value: 1}, {Key: "created_at", Value: -1}}},
	}
	_, err := col.Indexes().CreateMany(ctx, indexes)
	return err
}

func ensureCommunityIndexes(ctx context.Context, posts *mongo.Collection, comments *mongo.Collection) error {
	postIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "channel_id", Value: 1}, {Key: "created_at", Value: -1}}},
	}
	if _, err := posts.Indexes().CreateMany(ctx, postIndexes); err != nil {
		return err
	}

	commentIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "post_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
	}
	_, err := comments.Indexes().CreateMany(ctx, commentIndexes)
	return err
}

func ensureViewingIndexes(ctx context.Context, col *mongo.Collection) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "video_id", Value: 1}}},
	}
	_, err := col.Indexes().CreateMany(ctx, indexes)
	return err
}
