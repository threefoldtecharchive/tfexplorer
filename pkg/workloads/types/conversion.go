package types

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// ConversionCollection name
	ConversionCollection = "conversion"
)

// ErrErrNoConversion means no conversion is saved for a user
var ErrNoConversion = errors.New("no conversion yet")

type ConversionDoc struct {
	User      schema.ID        `bson:"user"`
	Workloads []WorkloaderType `bson:"workloads"`
	Converted bool             `bson:"converted"`
	Timestamp int64            `bson:"timestamp"`
}

// SaveUserConversion saves a conversion for a user
func SaveUserConversion(ctx context.Context, db *mongo.Database, user schema.ID, workloads []WorkloaderType) error {
	cd := ConversionDoc{
		User:      user,
		Workloads: workloads,
		Converted: false,
		Timestamp: time.Now().Unix(),
	}
	_, err := db.Collection(ConversionCollection).InsertOne(ctx, cd)
	return errors.Wrap(err, "could not insert conversion doc")
}

// GetUserConversion loads a conversion for a user
func GetUserConversion(ctx context.Context, db *mongo.Database, user schema.ID) (ConversionDoc, error) {
	var cd ConversionDoc
	res := db.Collection(ConversionCollection).FindOne(ctx, bson.M{"user": user})
	err := res.Err()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return cd, ErrNoConversion
		}
		return cd, errors.Wrap(err, "could not load conversion doc")
	}
	if err = res.Decode(&cd); err != nil {
		return cd, errors.Wrap(err, "could not decode conversion doc")
	}
	return cd, nil
}

// SetUserConversionSucceeded updates the converted field
func SetUserConversionSucceeded(ctx context.Context, db *mongo.Database, user schema.ID) error {
	_, err := db.Collection(ConversionCollection).UpdateOne(ctx, bson.M{"user": user}, bson.M{"$set": bson.M{"converted": true}})
	return errors.Wrap(err, "could not update converted field")
}
