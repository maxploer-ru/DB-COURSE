package migrate

import (
	"context"
	"fmt"

	mongoinfra "ZVideo/internal/infrastructure/db/mongo"
	"ZVideo/internal/infrastructure/db/mongo/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DefaultMigrations() []Migration {
	return []Migration{
		{
			ID:       "001_init_indexes",
			Checksum: "001_init_indexes_v1",
			Up: func(ctx context.Context, db *mongo.Database) error {
				return mongoinfra.EnsureIndexes(ctx, db)
			},
			Down: nil,
		},
		{
			ID:       "002_seed_roles",
			Checksum: "002_seed_roles_v1",
			Up:       seedRolesUp,
			Down:     seedRolesDown,
		},
	}
}

func seedRolesUp(ctx context.Context, db *mongo.Database) error {
	roles := []models.Role{
		{ID: 1, Name: "admin", IsDefault: false},
		{ID: 2, Name: "moderator", IsDefault: false},
		{ID: 3, Name: "user", IsDefault: true},
	}
	col := db.Collection(mongoinfra.CollectionRoles)
	opts := options.Update().SetUpsert(true)

	for _, role := range roles {
		filter := bson.M{"name": role.Name}
		update := bson.M{"$setOnInsert": bson.M{
			"_id":        role.ID,
			"name":       role.Name,
			"is_default": role.IsDefault,
		}}
		if _, err := col.UpdateOne(ctx, filter, update, opts); err != nil {
			return fmt.Errorf("seed role %s: %w", role.Name, err)
		}
	}
	return nil
}

func seedRolesDown(ctx context.Context, db *mongo.Database) error {
	_, err := db.Collection(mongoinfra.CollectionRoles).DeleteMany(ctx, bson.M{"name": bson.M{"$in": []string{"admin", "moderator", "user"}}})
	return err
}
