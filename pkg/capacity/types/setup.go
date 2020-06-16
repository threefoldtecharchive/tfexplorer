package types

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Setup sets up indexes for types, must be called at least
// Onetime during the life time of the object
func Setup(ctx context.Context, db *mongo.Database) error {
	col := db.Collection(CapacityPoolCollection)

	indexes := []mongo.IndexModel{
		{
			Keys: bson.M{"customer_tid": 1},
		},
	}

	if _, err := col.Indexes().CreateMany(ctx, indexes); err != nil {
		return err
	}

	return nil
}
