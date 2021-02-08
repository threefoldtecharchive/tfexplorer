package types

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/models"
	generated "github.com/threefoldtech/tfexplorer/models/generated/directory"

	"github.com/threefoldtech/tfexplorer/mw"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// FarmCollection db collection name
	FarmThreebotPriceCollection = "farmthreebotprice"
)

//Farm mongo db wrapper for generated TfgridDirectoryFarm
type FarmThreebotPrice generated.FarmThreebotPrice

// Validate validates farm object
func (f *FarmThreebotPrice) Validate() error {
	if f.ThreebotId == 0 {
		return fmt.Errorf("threebot_id is required")
	}
	if f.FarmId == 0 {
		return fmt.Errorf("farm_id is required")
	}

	return nil
}

// FarmQuery helper to parse query string
type FarmThreebotPriceQuery struct {
	FarmID     int64
	ThreebotID int64
}

// Parse querystring from request
func (f *FarmThreebotPriceQuery) Parse(r *http.Request) mw.Response {
	var err error
	f.ThreebotID, err = models.QueryInt(r, "threebot_id")
	if err != nil {
		return mw.BadRequest(errors.Wrap(err, "threebot_id should be a integer"))
	}
	f.FarmID, err = models.QueryInt(r, "farm_id")
	if err != nil {
		return mw.BadRequest(errors.Wrap(err, "farm_id should be a integer"))
	}
	return nil
}

// FarmThreebotPriceFilter type
type FarmThreebotPriceFilter bson.D

// WithFarmID filter farm with ID
func (f FarmThreebotPriceFilter) WithFarmID(id int64) FarmThreebotPriceFilter {
	return append(f, bson.E{Key: "farm_id", Value: id})
}

// WithThreebotID filter threebot with ID
func (f FarmThreebotPriceFilter) WithThreebotID(id int64) FarmThreebotPriceFilter {
	return append(f, bson.E{Key: "threebot_id", Value: id})
}

// Find run the filter and return a cursor result
func (f FarmThreebotPriceFilter) Find(ctx context.Context, db *mongo.Database, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	col := db.Collection(FarmThreebotPriceCollection)
	if f == nil {
		f = FarmThreebotPriceFilter{}
	}

	return col.Find(ctx, f, opts...)
}

// Count number of documents matching
func (f FarmThreebotPriceFilter) Count(ctx context.Context, db *mongo.Database) (int64, error) {
	col := db.Collection(FarmCollection)
	if f == nil {
		f = FarmThreebotPriceFilter{}
	}

	return col.CountDocuments(ctx, f)
}

// Get one farm that matches the filter
func (f FarmThreebotPriceFilter) Get(ctx context.Context, db *mongo.Database) (farmThreebotPrice FarmThreebotPrice, err error) {
	if f == nil {
		f = FarmThreebotPriceFilter{}
	}
	col := db.Collection(FarmThreebotPriceCollection)
	result := col.FindOne(ctx, f, options.FindOne())

	err = result.Err()
	if err != nil {
		return
	}

	err = result.Decode(&farmThreebotPrice)
	return
}

// Delete deletes one farm that match the filter
func (f FarmThreebotPriceFilter) Delete(ctx context.Context, db *mongo.Database) (err error) {
	if f == nil {
		f = FarmThreebotPriceFilter{}
	}
	col := db.Collection(FarmThreebotPriceCollection)
	_, err = col.DeleteOne(ctx, f, options.Delete())
	return err
}

// FarmThreebotPriceCreate creates a new farm threebot price
func FarmThreebotPriceCreateOrUpdate(ctx context.Context, db *mongo.Database, farmThreebotPrice FarmThreebotPrice) error {
	if err := farmThreebotPrice.Validate(); err != nil {
		return err
	}
	// update is a subset of Farm that only has the updatable fields.
	// this to preven the farmer from overriding other managed fields
	// like the list of IPs

	update := struct {
		ThreebotId           int64                        `bson:"threebot_id" json:"threebot_id"`
		FarmId               int64                        `bson:"farm_id" json:"farm_id"`
		CustomCloudUnitPrice generated.NodeCloudUnitPrice `bson:"custom_cloud_unit_price" json:"custom_cloud_unit_price"`
	}{
		ThreebotId:           farmThreebotPrice.ThreebotId,
		FarmId:               farmThreebotPrice.FarmId,
		CustomCloudUnitPrice: farmThreebotPrice.CustomCloudUnitPrice,
	}
	opts := options.Update().SetUpsert(true)

	col := db.Collection(FarmThreebotPriceCollection)
	f := FarmThreebotPriceFilter{}.WithFarmID(farmThreebotPrice.FarmId).WithThreebotID(farmThreebotPrice.ThreebotId)
	_, err := col.UpdateOne(ctx, f, bson.M{"$set": update}, opts)
	return err
}
