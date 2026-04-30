package migrate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	migrationsCollection = "schema_migrations"
	lockCollection       = "migration_locks"
	lockID               = "mongo_migrations"
	defaultLockTTL       = 2 * time.Minute
)

type Migration struct {
	ID       string
	Checksum string
	Up       func(ctx context.Context, db *mongo.Database) error
	Down     func(ctx context.Context, db *mongo.Database) error
}

type appliedMigration struct {
	ID        string    `bson:"_id"`
	Checksum  string    `bson:"checksum"`
	AppliedAt time.Time `bson:"applied_at"`
}

func Apply(ctx context.Context, db *mongo.Database, direction string, steps int, migrations []Migration) error {
	if direction != "up" && direction != "down" {
		return fmt.Errorf("invalid direction: %s", direction)
	}

	if err := ensureLockIndex(ctx, db); err != nil {
		return err
	}
	release, err := acquireLock(ctx, db, defaultLockTTL)
	if err != nil {
		return err
	}
	defer func() {
		_ = release(ctx, db)
	}()

	if err := ensureMigrationsIndex(ctx, db); err != nil {
		return err
	}

	migrationByID := make(map[string]Migration)
	for _, migration := range migrations {
		if migration.ID == "" {
			return errors.New("migration ID cannot be empty")
		}
		if _, exists := migrationByID[migration.ID]; exists {
			return fmt.Errorf("duplicate migration ID: %s", migration.ID)
		}
		migrationByID[migration.ID] = migration
	}

	sorted := make([]Migration, 0, len(migrations))
	sorted = append(sorted, migrations...)
	sortMigrations(sorted)

	applied, err := listApplied(ctx, db)
	if err != nil {
		return err
	}

	if direction == "up" {
		return applyUp(ctx, db, sorted, applied, steps)
	}
	return applyDown(ctx, db, migrationByID, steps)
}

func applyUp(ctx context.Context, db *mongo.Database, migrations []Migration, applied map[string]appliedMigration, steps int) error {
	var pending []Migration
	for _, migration := range migrations {
		existing, ok := applied[migration.ID]
		if ok {
			if existing.Checksum != checksum(migration) {
				return fmt.Errorf("checksum mismatch for %s", migration.ID)
			}
			continue
		}
		pending = append(pending, migration)
	}

	if steps > 0 && steps < len(pending) {
		pending = pending[:steps]
	}
	if len(pending) == 0 {
		return nil
	}

	for _, migration := range pending {
		if migration.Up == nil {
			return fmt.Errorf("missing up migration for %s", migration.ID)
		}
		if err := migration.Up(ctx, db); err != nil {
			return fmt.Errorf("apply %s: %w", migration.ID, err)
		}
		record := appliedMigration{
			ID:        migration.ID,
			Checksum:  checksum(migration),
			AppliedAt: time.Now().UTC(),
		}
		if _, err := db.Collection(migrationsCollection).InsertOne(ctx, record); err != nil {
			return fmt.Errorf("record migration %s: %w", migration.ID, err)
		}
	}
	return nil
}

func applyDown(ctx context.Context, db *mongo.Database, migrationByID map[string]Migration, steps int) error {
	applied, err := listAppliedOrdered(ctx, db, -1)
	if err != nil {
		return err
	}
	if steps > 0 && steps < len(applied) {
		applied = applied[:steps]
	}
	if len(applied) == 0 {
		return nil
	}

	for _, item := range applied {
		migration, ok := migrationByID[item.ID]
		if !ok {
			return fmt.Errorf("missing migration definition for %s", item.ID)
		}
		if migration.Down == nil {
			return fmt.Errorf("migration %s is not reversible", item.ID)
		}
		if err := migration.Down(ctx, db); err != nil {
			return fmt.Errorf("rollback %s: %w", item.ID, err)
		}
		if _, err := db.Collection(migrationsCollection).DeleteOne(ctx, bson.M{"_id": item.ID}); err != nil {
			return fmt.Errorf("remove migration %s: %w", item.ID, err)
		}
	}
	return nil
}

func listApplied(ctx context.Context, db *mongo.Database) (map[string]appliedMigration, error) {
	cursor, err := db.Collection(migrationsCollection).Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}
	defer cursor.Close(ctx)

	result := make(map[string]appliedMigration)
	for cursor.Next(ctx) {
		var item appliedMigration
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		result[item.ID] = item
	}
	return result, cursor.Err()
}

func listAppliedOrdered(ctx context.Context, db *mongo.Database, order int) ([]appliedMigration, error) {
	opts := options.Find().SetSort(bson.D{{Key: "applied_at", Value: order}})
	cursor, err := db.Collection(migrationsCollection).Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}
	defer cursor.Close(ctx)

	var result []appliedMigration
	for cursor.Next(ctx) {
		var item appliedMigration
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, cursor.Err()
}

func checksum(migration Migration) string {
	if migration.Checksum != "" {
		return migration.Checksum
	}
	return migration.ID
}

func sortMigrations(migrations []Migration) {
	for i := 0; i < len(migrations)-1; i++ {
		for j := i + 1; j < len(migrations); j++ {
			if migrations[i].ID > migrations[j].ID {
				migrations[i], migrations[j] = migrations[j], migrations[i]
			}
		}
	}
}

func ensureMigrationsIndex(ctx context.Context, db *mongo.Database) error {
	index := mongo.IndexModel{
		Keys: bson.D{{Key: "applied_at", Value: 1}},
		Options: options.Index().
			SetName("schema_migrations_applied_at"),
	}
	_, err := db.Collection(migrationsCollection).Indexes().CreateOne(ctx, index)
	return err
}

func ensureLockIndex(ctx context.Context, db *mongo.Database) error {
	index := mongo.IndexModel{
		Keys: bson.D{{Key: "locked_until", Value: 1}},
		Options: options.Index().
			SetExpireAfterSeconds(0).
			SetName("migration_locks_ttl"),
	}
	_, err := db.Collection(lockCollection).Indexes().CreateOne(ctx, index)
	return err
}

func acquireLock(ctx context.Context, db *mongo.Database, ttl time.Duration) (func(context.Context, *mongo.Database) error, error) {
	now := time.Now().UTC()
	lockedUntil := now.Add(ttl)
	owner, _ := os.Hostname()

	filter := bson.M{
		"_id": lockID,
		"$or": []bson.M{
			{"locked_until": bson.M{"$lt": now}},
			{"locked_until": bson.M{"$exists": false}},
		},
	}
	update := bson.M{
		"$set": bson.M{
			"locked_at":    now,
			"locked_until": lockedUntil,
			"owner":        owner,
		},
	}
	opts := options.Update().SetUpsert(true)
	res, err := db.Collection(lockCollection).UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, fmt.Errorf("acquire mongo migration lock: %w", err)
	}
	if res.MatchedCount == 0 && res.UpsertedCount == 0 {
		return nil, errors.New("mongo migration lock already held")
	}

	release := func(ctx context.Context, db *mongo.Database) error {
		_, err := db.Collection(lockCollection).DeleteOne(ctx, bson.M{"_id": lockID, "owner": owner})
		return err
	}
	return release, nil
}
