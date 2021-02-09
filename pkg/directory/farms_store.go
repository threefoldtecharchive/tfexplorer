package directory

import (
	"context"
	"net"

	"github.com/pkg/errors"
	generated "github.com/threefoldtech/tfexplorer/models/generated/directory"
	directory "github.com/threefoldtech/tfexplorer/pkg/directory/types"

	"github.com/threefoldtech/tfexplorer/schema"
	"github.com/zaibon/httpsig"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FarmAPI holds farm releated handlers
type FarmAPI struct {
	verifier *httpsig.Verifier
}

// List farms
// TODO: add paging arguments
func (s *FarmAPI) List(ctx context.Context, db *mongo.Database, filter directory.FarmFilter, opts ...*options.FindOptions) ([]directory.Farm, int64, error) {

	cur, err := filter.Find(ctx, db, opts...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to list farms")
	}
	defer cur.Close(ctx)
	out := []directory.Farm{}
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, errors.Wrap(err, "failed to load farm list")
	}

	count, err := filter.Count(ctx, db)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to count entries in farms collection")
	}

	return out, count, nil
}

// GetByName gets a farm by name
func (s *FarmAPI) GetByName(ctx context.Context, db *mongo.Database, name string) (directory.Farm, error) {
	var filter directory.FarmFilter
	filter = filter.WithName(name)

	return filter.Get(ctx, db)
}

// GetByID gets a farm by ID
func (s *FarmAPI) GetByID(ctx context.Context, db *mongo.Database, id int64) (directory.Farm, error) {
	var filter directory.FarmFilter
	filter = filter.WithID(schema.ID(id))

	return filter.Get(ctx, db)
}

// Add add farm to store
func (s *FarmAPI) Add(ctx context.Context, db *mongo.Database, farm directory.Farm) (schema.ID, error) {
	return directory.FarmCreate(ctx, db, farm)
}

// Update farm information
func (s *FarmAPI) Update(ctx context.Context, db *mongo.Database, id schema.ID, farm directory.Farm) error {
	return directory.FarmUpdate(ctx, db, id, farm)
}

// PushIP push ip
func (s *FarmAPI) PushIP(ctx context.Context, db *mongo.Database, id schema.ID, ip schema.IPCidr, gw net.IP) error {
	return directory.FarmPushIP(ctx, db, id, ip, gw)
}

// RemoveIP removes ip
func (s *FarmAPI) RemoveIP(ctx context.Context, db *mongo.Database, id schema.ID, ip schema.IPCidr) error {
	return directory.FarmRemoveIP(ctx, db, id, ip)
}

// Delete deletes a farm by ID
func (s FarmAPI) Delete(ctx context.Context, db *mongo.Database, id int64) error {
	var filter directory.FarmFilter
	filter = filter.WithID(schema.ID(id))
	return filter.Delete(ctx, db)
}

func (s *FarmAPI) GetFarmCustomPrices(ctx context.Context, db *mongo.Database, farmId int64) ([]directory.FarmThreebotPrice, int64, error) {
	var filter directory.FarmThreebotPriceFilter
	filter = filter.WithFarmID(farmId)
	var count int64

	cur, err := filter.Find(ctx, db)

	if err != nil {
		return nil, count, errors.Wrap(err, "failed to list farmthreebotprice")
	}
	defer cur.Close(ctx)
	out := []directory.FarmThreebotPrice{}
	if err := cur.All(ctx, &out); err != nil {
		return nil, count, errors.Wrap(err, "failed to load farmthreebotprice list")
	}

	count, err = filter.Count(ctx, db)
	if err != nil {
		return nil, count, errors.Wrap(err, "failed to count entries in farms collection")
	}

	return out, count, nil
}

func (s *FarmAPI) GetFarmCustomPriceForThreebot(ctx context.Context, db *mongo.Database, farmId, threebotId int64) (directory.FarmThreebotPrice, error) {
	var filter directory.FarmThreebotPriceFilter
	filter = filter.WithFarmID(farmId).WithThreebotID(threebotId)
	farmThreebotPrice, err := filter.Get(ctx, db)
	if err != nil {
		// check the default pricing or return the explorer pricing..
		farm, err := s.GetByID(ctx, db, farmId)
		if err != nil {
			return directory.FarmThreebotPrice{}, errors.Wrap(err, "failed to find farm") //todo add farm id..
		}
		if farm.EnableCustomPricing {
			// is there a better way to unwrap the returned farm?
			unwrappedFromMongoFarmPrice := generated.NodeCloudUnitPrice{}
			unwrappedFromMongoFarmPrice.CU = farm.DefaultCloudUnitPrice.CU
			unwrappedFromMongoFarmPrice.SU = farm.DefaultCloudUnitPrice.SU
			unwrappedFromMongoFarmPrice.NU = farm.DefaultCloudUnitPrice.NU
			return directory.FarmThreebotPrice{FarmId: farmId, ThreebotId: threebotId, CustomCloudUnitPrice: unwrappedFromMongoFarmPrice}, nil
		}

		return directory.FarmThreebotPrice{}, errors.Wrap(err, "farmer doesn't use custom pricing. should fallback to explorer generic calculation")
	}
	return farmThreebotPrice, nil
}
