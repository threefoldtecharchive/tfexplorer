package types

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/models"
	model "github.com/threefoldtech/tfexplorer/models/workloads"
	"github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// WorkloadCollection db collection name
	WorkloadCollection = "workload"
)

// ApplyQueryFilterWorkload parses the query string
func ApplyQueryFilterWorkload(r *http.Request, filter WorkloadFilter) (WorkloadFilter, error) {
	var err error
	customerid, err := models.QueryInt(r, "customer_tid")
	if err != nil {
		return nil, errors.Wrap(err, "customer_tid should be an integer")
	}
	if customerid != 0 {
		filter = filter.WithCustomerID(int(customerid))
	}
	sNextAction := r.FormValue("next_action")
	if len(sNextAction) != 0 {
		nextAction, err := strconv.ParseInt(sNextAction, 10, 0)
		if err != nil {
			return nil, errors.Wrap(err, "next_action should be an integer")
		}
		filter = filter.WithNextAction(model.NextActionEnum(nextAction))
	}
	return filter, nil
}

// WorkloadFilter type
type WorkloadFilter bson.D

// WithID filter workload with ID
func (f WorkloadFilter) WithID(id schema.ID) WorkloadFilter {
	return append(f, bson.E{Key: "_id", Value: id})
}

// WithIDGE return find workloads with
func (f WorkloadFilter) WithIDGE(id schema.ID) WorkloadFilter {
	return append(f, bson.E{
		Key: "_id", Value: bson.M{"$gte": id},
	})
}

// WithNextAction filter workloads with next action
func (f WorkloadFilter) WithNextAction(action model.NextActionEnum) WorkloadFilter {
	return append(f, bson.E{
		Key: "next_action", Value: action,
	})
}

// WithCustomerID filter workload on customer
func (f WorkloadFilter) WithCustomerID(customerID int) WorkloadFilter {
	return append(f, bson.E{
		Key: "customer_tid", Value: customerID,
	})
}

// WithNodeID searsch workloads with NodeID
func (f WorkloadFilter) WithNodeID(id string) WorkloadFilter {
	return append(f, bson.E{
		Key: "node_id", Value: id,
	})
}

// WithReference searches workloads with reference
func (f WorkloadFilter) WithReference(ref string) WorkloadFilter {
	return append(f, bson.E{
		Key: "reference", Value: ref,
	})
}

// Or returns filter that reads as (f or o)
func (f WorkloadFilter) Or(o WorkloadFilter) WorkloadFilter {
	return WorkloadFilter{
		bson.E{
			Key:   "$or",
			Value: bson.A{f, o},
		},
	}
}

// Get gets single workload that matches the filter
func (f WorkloadFilter) Get(ctx context.Context, db *mongo.Database) (model.Workloader, error) {
	if f == nil {
		f = WorkloadFilter{}
	}
	var w WorkloaderCodec

	result := db.Collection(WorkloadCollection).FindOne(ctx, f)
	if err := result.Err(); err != nil {
		return w, err
	}

	if err := result.Decode(&w); err != nil {
		return w, errors.Wrap(err, "could not decode workload type")
	}

	return w, nil
}

// Find all users workloads matches filter
func (f WorkloadFilter) Find(ctx context.Context, db *mongo.Database, opts ...*options.FindOptions) ([]model.Workloader, error) {
	if f == nil {
		f = WorkloadFilter{}
	}

	cursor, err := db.Collection(WorkloadCollection).Find(ctx, f, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get workload cursor")
	}

	ws := []model.Workloader{}

	for cursor.Next(ctx) {
		var w WorkloaderCodec
		if err = cursor.Decode(&w); err != nil {
			return nil, errors.Wrap(err, "could not decode workload type")
		}
		ws = append(ws, w)
	}

	return ws, nil
}

// FindCursor all users workloads matches filter, and return a plain cursor to them
func (f WorkloadFilter) FindCursor(ctx context.Context, db *mongo.Database, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	if f == nil {
		f = WorkloadFilter{}
	}

	cursor, err := db.Collection(WorkloadCollection).Find(ctx, f, opts...)

	return cursor, errors.Wrap(err, "failed to get workload cursor")
}

// Count number of documents matching
func (f WorkloadFilter) Count(ctx context.Context, db *mongo.Database) (int64, error) {
	col := db.Collection(WorkloadCollection)
	if f == nil {
		f = WorkloadFilter{}
	}

	return col.CountDocuments(ctx, f)
}

// WithPoolID searches for workloads with pool id
func (f WorkloadFilter) WithPoolID(poolID int64) WorkloadFilter {
	return append(f, bson.E{
		Key: "pool_id", Value: poolID,
	})
}

// WorkloadCreate save new workload to database.
// NOTE: use reservations only that are returned from calling Pipeline.Next()
// no validation is done here, this is just a CRUD operation
func WorkloadCreate(ctx context.Context, db *mongo.Database, w model.Workloader) (schema.ID, error) {
	id := models.MustID(ctx, db, ReservationCollection)
	w.Contract().ID = id

	_, err := db.Collection(WorkloadCollection).InsertOne(ctx, w)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// WorkloadsLastID get the current last ID number in the workloads collection
func WorkloadsLastID(ctx context.Context, db *mongo.Database) (schema.ID, error) {
	return models.LastID(ctx, db, ReservationCollection)
}

// WorkloadSetNextAction update the workload next action in db
func WorkloadSetNextAction(ctx context.Context, db *mongo.Database, id schema.ID, action model.NextActionEnum) error {
	var filter WorkloadFilter
	filter = filter.WithID(id)

	col := db.Collection(WorkloadCollection)
	_, err := col.UpdateOne(ctx, filter, bson.M{
		"$set": bson.M{
			"next_action": action,
		},
	})

	if err != nil {
		return err
	}

	return nil
}

// WorkloadToDeploy marks a workload to deploy and schedule it for the nodes
// it's a short cut to SetNextAction then PushWorkloads
func WorkloadToDeploy(ctx context.Context, db *mongo.Database, w model.Workloader) error {
	// update workload
	if err := WorkloadSetNextAction(ctx, db, w.Contract().ID, Deploy); err != nil {
		return errors.Wrap(err, "failed to set workload to DEPLOY state")
	}

	//queue for processing
	if err := WorkloadTypePush(ctx, db, w); err != nil {
		return errors.Wrap(err, "failed to schedule workload for deploying")
	}

	return nil
}

//WorkloadPushSignature push signature to workload
func WorkloadPushSignature(ctx context.Context, db *mongo.Database, id schema.ID, mode SignatureMode, signature model.SigningSignature) error {
	// this function just push the signature to the reservation array
	// there are not other checks involved here. So before calling this function
	// we need to ensure the signature has the rights to be pushed
	var filter WorkloadFilter
	filter = filter.WithID(id)
	col := db.Collection(WorkloadCollection)
	_, err := col.UpdateOne(ctx, filter, bson.M{
		"$push": bson.M{
			string(mode): signature,
		},
	})

	return err
}

// WorkloadTypePush pushes a workload to the queue
func WorkloadTypePush(ctx context.Context, db *mongo.Database, w model.Workloader) error {
	col := db.Collection(queueCollection)
	_, err := col.InsertOne(ctx, w)

	return err
}

// WorkloadTypePop removes workload from queue
func WorkloadTypePop(ctx context.Context, db *mongo.Database, id string, nodeID string) error {
	col := db.Collection(queueCollection)
	_, err := col.DeleteOne(ctx, bson.M{"workload_id": id, "node_id": nodeID})

	return err
}

// WorkloadResultPush pushes result to a reservation result array.
// NOTE: this is just a crud operation, no validation is done here
func WorkloadResultPush(ctx context.Context, db *mongo.Database, id schema.ID, result model.Result) error {
	col := db.Collection(WorkloadCollection)
	var filter WorkloadFilter
	filter = filter.WithID(id)

	_, err := col.UpdateOne(ctx, filter,
		bson.M{
			"$set": bson.M{
				"result": result,
			},
		},
	)

	return err
}

// WorkloaderCodec is a struc used to encode/decode model.Workloads
type WorkloaderCodec struct {
	model.Workloader
}

// MarshalBSON implements bson.Marshaller
func (w WorkloaderCodec) MarshalBSON() ([]byte, error) {
	return bson.Marshal(w.Workloader)
}

// UnmarshalBSON implements bson.Unmarshaller
func (w *WorkloaderCodec) UnmarshalBSON(buf []byte) error {
	workload, err := model.UnmarshalBSON(buf)
	if err != nil {
		return err
	}

	if w == nil {
		w = &WorkloaderCodec{}
	}

	*w = WorkloaderCodec{Workloader: workload}

	return nil
}

// MarshalJSON implements JSON.Marshaller
func (w WorkloaderCodec) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.Workloader)
}

// UnmarshalJSON implements JSON.Unmarshaller
func (w *WorkloaderCodec) UnmarshalJSON(buf []byte) error {
	workload, err := model.UnmarshalJSON(buf)
	if err != nil {
		return err
	}

	if w == nil {
		w = &WorkloaderCodec{}
	}

	*w = WorkloaderCodec{Workloader: workload}

	return nil
}
