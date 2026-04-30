package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type counterDoc struct {
	ID  string `bson:"_id"`
	Seq int    `bson:"seq"`
}

func NextID(ctx context.Context, db *mongo.Database, name string) (int, error) {
	col := db.Collection(CollectionCounters)
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	update := bson.M{"$inc": bson.M{"seq": 1}}

	var doc counterDoc
	if err := col.FindOneAndUpdate(ctx, bson.M{"_id": name}, update, opts).Decode(&doc); err != nil {
		return 0, fmt.Errorf("get next id for %s: %w", name, err)
	}
	return doc.Seq, nil
}
