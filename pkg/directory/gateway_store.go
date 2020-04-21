package directory

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	generated "github.com/threefoldtech/tfexplorer/models/generated/directory"
	"github.com/threefoldtech/tfexplorer/mw"
	directory "github.com/threefoldtech/tfexplorer/pkg/directory/types"
	"github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GatewayAPI holds api for gateways
type GatewayAPI struct{}

type gatewayQuery struct {
	// FarmID  int64
	Country string
	City    string
	// CRU     int64
	// MRU     int64
	// SRU     int64
	// HRU     int64
	// Proofs  bool
}

func (n *gatewayQuery) Parse(r *http.Request) mw.Response {
	// n.FarmID, err = models.QueryInt(r, "farm")
	// if err != nil {
	// 	return mw.BadRequest(errors.Wrap(err, "invalid farm id"))
	// }
	n.Country = r.URL.Query().Get("country")
	n.City = r.URL.Query().Get("city")
	return nil
}

// List all gateways
func (s *GatewayAPI) List(ctx context.Context, db *mongo.Database, q gatewayQuery, opts ...*options.FindOptions) ([]directory.Gateway, int64, error) {
	var filter directory.GatewayFilter
	// if q.FarmID > 0 {
	// 	filter = filter.WithFarmID(schema.ID(q.FarmID))
	// }
	// filter = filter.WithTotalCap(q.CRU, q.MRU, q.HRU, q.SRU)
	filter = filter.WithLocation(q.Country, q.City)

	cur, err := filter.Find(ctx, db, opts...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to list nodes")
	}

	defer cur.Close(ctx)
	out := []directory.Gateway{}
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, errors.Wrap(err, "failed to load node list")
	}

	count, err := filter.Count(ctx, db)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to count entries in nodes collection")
	}

	return out, count, nil
}

// Get a single gateway
func (s *GatewayAPI) Get(ctx context.Context, db *mongo.Database, gwID string, includeProofs bool) (directory.Node, error) {
	var filter directory.GatewayFilter
	filter = filter.WithGWID(gwID)
	return filter.Get(ctx, db, includeProofs)
}

// Exists tests if node exists
func (s *GatewayAPI) Exists(ctx context.Context, db *mongo.Database, gwID string) (bool, error) {
	var filter directory.GatewayFilter
	filter = filter.WithGWID(gwID)

	count, err := filter.Count(ctx, db)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Count counts the number of document in the collection
func (s *GatewayAPI) Count(ctx context.Context, db *mongo.Database, filter directory.NodeFilter) (int64, error) {
	return filter.Count(ctx, db)
}

// Add a node to the store
func (s *GatewayAPI) Add(ctx context.Context, db *mongo.Database, gw directory.Gateway) (schema.ID, error) {
	return directory.GatewayCreate(ctx, db, gw)
}

// func (s *GatewayAPI) updateTotalCapacity(ctx context.Context, db *mongo.Database, gwID string, capacity generated.ResourceAmount) error {
// 	return directory.NodeUpdateTotalResources(ctx, db, gwID, capacity)
// }

func (s *GatewayAPI) updateReservedCapacity(ctx context.Context, db *mongo.Database, gwID string, capacity generated.ResourceAmount) error {
	return directory.GatewayUpdateReservedResources(ctx, db, gwID, capacity)
}

func (s *GatewayAPI) updateUptime(ctx context.Context, db *mongo.Database, gwID string, uptime int64) error {
	return directory.GatewayUpdateUptime(ctx, db, gwID, uptime)
}

// func (s *GatewayAPI) updateFreeToUse(ctx context.Context, db *mongo.Database, gwID string, freeToUse bool) error {
// 	return directory.NodeUpdateFreeToUse(ctx, db, gwID, freeToUse)
// }

func (s *GatewayAPI) updateWorkloadsAmount(ctx context.Context, db *mongo.Database, gwID string, workloads generated.WorkloadAmount) error {
	return directory.GatewayUpdateWorkloadsAmount(ctx, db, gwID, workloads)
}

// Requires is a wrapper that makes sure gateway with that key exists before
// running the handler
func (s *GatewayAPI) Requires(key string, handler mw.Action) mw.Action {
	return func(r *http.Request) (interface{}, mw.Response) {
		gwID, ok := mux.Vars(r)[key]
		if !ok {
			// programming error, we should panic in this case
			panic("invalid node-id key")
		}

		db := mw.Database(r)

		exists, err := s.Exists(r.Context(), db, gwID)
		if err != nil {
			return nil, mw.Error(err)
		} else if !exists {
			return nil, mw.NotFound(fmt.Errorf("gateway '%s' not found", gwID))
		}

		return handler(r)
	}
}
