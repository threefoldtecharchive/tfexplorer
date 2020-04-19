package types

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jbenet/go-base58"
	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/models"
	generated "github.com/threefoldtech/tfexplorer/models/generated/directory"
	"github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// GatewayCollection db collection name
	GatewayCollection = "gateway"
)

// Gateway model
type Gateway generated.Gateway

// Validate node
func (n *Gateway) Validate() error {
	if len(n.NodeId) == 0 {
		return fmt.Errorf("node_is is required")
	}

	if len(n.OsVersion) == 0 {
		return fmt.Errorf("os_version is required")
	}

	if len(n.PublicKeyHex) == 0 {
		return fmt.Errorf("public_key_hex is required")
	}

	pk, err := hex.DecodeString(n.PublicKeyHex)
	if err != nil {
		return errors.Wrap(err, "fail to decode public key")
	}

	if n.NodeId != base58.Encode(pk) {
		return fmt.Errorf("nodeID and public key does not match")
	}

	// Unfortunately, jsx schema does not support nil types
	// so this is the only way to check if values are not set
	empty := generated.Location{}
	if n.Location == empty {
		return fmt.Errorf("location is required")
	}

	return nil
}

// GatewayFilter type
type GatewayFilter bson.D

// WithID filter node with ID
func (f GatewayFilter) WithID(id schema.ID) GatewayFilter {
	return append(f, bson.E{Key: "_id", Value: id})
}

// WithGWID search nodes with this node id
func (f GatewayFilter) WithGWID(id string) GatewayFilter {
	return append(f, bson.E{Key: "node_id", Value: id})
}

// WithLocation search the nodes that are located in country and or city
func (f GatewayFilter) WithLocation(country, city string) GatewayFilter {
	if country != "" {
		f = append(f, bson.E{Key: "location.country", Value: country})
	}
	if city != "" {
		f = append(f, bson.E{Key: "location.city", Value: city})
	}

	return f
}

// Find run the filter and return a cursor result
func (f GatewayFilter) Find(ctx context.Context, db *mongo.Database, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	col := db.Collection(GatewayCollection)
	if f == nil {
		f = GatewayFilter{}
	}

	return col.Find(ctx, f, opts...)
}

// Get one farm that matches the filter
func (f GatewayFilter) Get(ctx context.Context, db *mongo.Database, includeproofs bool) (node Node, err error) {
	if f == nil {
		f = GatewayFilter{}
	}

	col := db.Collection(GatewayCollection)

	result := col.FindOne(ctx, f)

	err = result.Err()
	if err != nil {
		return
	}

	err = result.Decode(&node)
	return
}

// Count number of documents matching
func (f GatewayFilter) Count(ctx context.Context, db *mongo.Database) (int64, error) {
	col := db.Collection(GatewayCollection)
	if f == nil {
		f = GatewayFilter{}
	}

	return col.CountDocuments(ctx, f)
}

// Delete deletes a node by ID
func (f GatewayFilter) Delete(ctx context.Context, db *mongo.Database) error {
	col := db.Collection(GatewayCollection)
	if f == nil {
		f = GatewayFilter{}
	}
	_, err := col.DeleteOne(ctx, f, options.Delete())
	return err
}

// GatewayCreate creates a new gateway
func GatewayCreate(ctx context.Context, db *mongo.Database, gw Gateway) (schema.ID, error) {
	if err := gw.Validate(); err != nil {
		return 0, err
	}

	var filter GatewayFilter
	filter = filter.WithGWID(gw.NodeId)

	var id schema.ID
	current, err := filter.Get(ctx, db, false)
	if err != nil {
		//TODO: check that this is a NOT FOUND error
		id, err = models.NextID(ctx, db, GatewayCollection)
		if err != nil {
			return id, err
		}
		gw.Created = schema.Date{Time: time.Now()}
	} else {
		id = current.ID
		// make sure we do NOT overwrite these field
		gw.Created = current.Created
		// gw.FreeToUse = current.FreeToUse
	}

	gw.ID = id
	gw.Updated = schema.Date{Time: time.Now()}

	col := db.Collection(GatewayCollection)
	_, err = col.UpdateOne(ctx, filter, bson.M{"$set": gw}, options.Update().SetUpsert(true))
	return id, err
}

func gwUpdate(ctx context.Context, db *mongo.Database, nodeID string, value interface{}) error {
	if nodeID == "" {
		return fmt.Errorf("invalid node id")
	}

	col := db.Collection(GatewayCollection)
	var filter GatewayFilter
	filter = filter.WithGWID(nodeID)
	_, err := col.UpdateOne(ctx, filter, bson.M{
		"$set": value,
	})

	return err
}

// // gwUpdateTotalResources sets the node total resources
// func gwUpdateTotalResources(ctx context.Context, db *mongo.Database, nodeID string, capacity generated.ResourceAmount) error {
// 	return gwUpdate(ctx, db, nodeID, bson.M{"total_resources": capacity})
// }

// GatewayUpdateReservedResources sets the node reserved resources
func GatewayUpdateReservedResources(ctx context.Context, db *mongo.Database, nodeID string, capacity generated.ResourceAmount) error {
	return gwUpdate(ctx, db, nodeID, bson.M{"reserved_resources": capacity})
}

// GatewayUpdateWorkloadsAmount sets the node reserved resources
func GatewayUpdateWorkloadsAmount(ctx context.Context, db *mongo.Database, nodeID string, workloads generated.GatewayResourceWorkloads) error {
	return gwUpdate(ctx, db, nodeID, bson.M{"workloads": workloads})
}

// GatewayUpdateUptime updates node uptime
func GatewayUpdateUptime(ctx context.Context, db *mongo.Database, nodeID string, uptime int64) error {
	return gwUpdate(ctx, db, nodeID, bson.M{
		"uptime":  uptime,
		"updated": schema.Date{Time: time.Now()},
	})
}
