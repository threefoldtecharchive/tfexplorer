package types

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/models"
	generated "github.com/threefoldtech/tfexplorer/models/generated/workloads"
	"github.com/threefoldtech/tfexplorer/schema"
	"github.com/threefoldtech/zos/pkg/crypto"
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
		filter = filter.WithNextAction(generated.NextActionEnum(nextAction))
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
func (f WorkloadFilter) WithNextAction(action generated.NextActionEnum) WorkloadFilter {
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
	//data_reservation.{containers, volumes, zdbs, networks, kubernetes}.node_id
	// we need to search ALL types for any workload that has the node ID
	or := []bson.M{}

	for _, typ := range []string{"containers", "volumes", "zdbs", "kubernetes", "proxies", "reverse_proxies", "subdomains", "domain_delegates", "gateway4to6", "capacity_pool"} {
		key := fmt.Sprintf("data_reservation.%s.node_id", typ)
		or = append(or, bson.M{key: id})
	}

	// network workload is special because node id is set on the network_resources.
	or = append(or, bson.M{"data_reservation.networks.network_resources.node_id": id})

	// we find any workload that has this node ID set.
	return append(f, bson.E{Key: "$or", Value: or})
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
func (f WorkloadFilter) Get(ctx context.Context, db *mongo.Database) (workload WorkloaderType, err error) {
	if f == nil {
		f = WorkloadFilter{}
	}

	result := db.Collection(ReservationCollection).FindOne(ctx, f)
	if err = result.Err(); err != nil {
		return
	}

	err = result.Decode(&workload)
	return
}

// Find all users that matches filter
func (f WorkloadFilter) Find(ctx context.Context, db *mongo.Database, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	if f == nil {
		f = WorkloadFilter{}
	}
	return db.Collection(ReservationCollection).Find(ctx, f, opts...)
}

// Count number of documents matching
func (f WorkloadFilter) Count(ctx context.Context, db *mongo.Database) (int64, error) {
	col := db.Collection(ReservationCollection)
	if f == nil {
		f = WorkloadFilter{}
	}

	return col.CountDocuments(ctx, f)
}

// WorkloaderType is a wrapper struct around the Workloader interface
type WorkloaderType struct {
	generated.Workloader
}

// Verify signature against Workload.JSON
// pk is the public key used as verification key in hex encoded format
// the signature is the signature to verify (in raw binary format)
func (w *WorkloaderType) Verify(pk string, sig []byte) error {
	key, err := crypto.KeyFromHex(pk)
	if err != nil {
		return errors.Wrap(err, "invalid verification key")
	}

	return crypto.Verify(key, []byte(w.GetJson()), sig)
}

// SignatureVerify is similar to Verify but the verification is done
// against `str(WorkloaderType.ID) + WorkloaderType.JSON`
func (w *WorkloaderType) SignatureVerify(pk string, sig []byte) error {
	key, err := crypto.KeyFromHex(pk)
	if err != nil {
		return errors.Wrap(err, "invalid verification key")
	}

	var buf bytes.Buffer
	if _, err := buf.WriteString(fmt.Sprint(int64(w.GetID()))); err != nil {
		return errors.Wrap(err, "failed to write id to buffer")
	}

	if _, err := buf.WriteString(w.GetJson()); err != nil {
		return errors.Wrap(err, "failed to write json to buffer")
	}

	return crypto.Verify(key, buf.Bytes(), sig)
}

// Expired checks if this reservation has expired
// func (w *WorkloaderType) Expired() bool {
// 	return time.Until(r.DataReservation.ExpirationReservation.Time) <= 0
// }

// IsAny checks if the workload status is any of the given status
func (w *WorkloaderType) IsAny(status ...generated.NextActionEnum) bool {
	for _, s := range status {
		if w.GetNextAction() == s {
			return true
		}
	}

	return false
}

//ResultOf return result of a workload ID
func (w *WorkloaderType) ResultOf(id string) *Result {
	if w.GetResult().WorkloadId == id {
		r := Result(w.GetResult())
		return &r
	}

	return nil
}

// AllDeleted checks of all workloads has been marked
func (w *WorkloaderType) AllDeleted() bool {
	result := w.GetResult()
	if result.State != generated.ResultStateDeleted {
		return false
	}

	return true
}

// IsSuccessfullyDeployed check if all the workloads defined in the reservation
// have sent a positive result
func (w *WorkloaderType) IsSuccessfullyDeployed() bool {
	succeeded := false
	if w.GetResult().State != generated.ResultStateOK {
		succeeded = false
	}
	return succeeded
}

// WorkloadCreate save new workload to database.
// NOTE: use reservations only that are returned from calling Pipeline.Next()
// no validation is done here, this is just a CRUD operation
func WorkloadCreate(ctx context.Context, db *mongo.Database, w WorkloaderType) (schema.ID, error) {
	id := models.MustID(ctx, db, ReservationCollection)
	w.SetID(id)

	_, err := db.Collection(WorkloadCollection).InsertOne(ctx, w)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// WorkloadsLastID get the current last ID number in the workloads collection
func WorkloadsLastID(ctx context.Context, db *mongo.Database) (schema.ID, error) {
	return models.LastID(ctx, db, WorkloadCollection)
}

// WorkloadSetNextAction update the workload next action in db
func WorkloadSetNextAction(ctx context.Context, db *mongo.Database, id schema.ID, action generated.NextActionEnum) error {
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
func WorkloadToDeploy(ctx context.Context, db *mongo.Database, w WorkloaderType) error {
	// update workload
	if err := WorkloadSetNextAction(ctx, db, w.GetID(), Deploy); err != nil {
		return errors.Wrap(err, "failed to set workload to DEPLOY state")
	}

	//queue for processing
	if err := WorkloadTypePush(ctx, db, w.Workload()); err != nil {
		return errors.Wrap(err, "failed to schedule workload for deploying")
	}

	return nil
}

//WorkloadPushSignature push signature to workload
func WorkloadPushSignature(ctx context.Context, db *mongo.Database, id schema.ID, mode SignatureMode, signature generated.SigningSignature) error {
	var filter WorkloadFilter
	filter = filter.WithID(id)
	col := db.Collection(WorkloadCollection)
	// NOTE: this should be a transaction not a bulk write
	// but i had so many issues with transaction, and i couldn't
	// get it to work. so I used bulk write in place instead
	// until we figure this issue out.
	// Note, the reason we don't just use addToSet is the signature
	// object always have the current 'time' which means it's a different
	// value than the one in the document even if it has same user id.
	_, err := col.BulkWrite(ctx, []mongo.WriteModel{
		mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(
			bson.M{
				"$pull": bson.M{
					string(mode): bson.M{"tid": signature.Tid},
				},
			}),
		mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(
			bson.M{
				"$addToSet": bson.M{
					string(mode): signature,
				},
			}),
	}, options.BulkWrite().SetOrdered(true))

	return err
}

// WorkloadTypePush pushes a workload to the queue
func WorkloadTypePush(ctx context.Context, db *mongo.Database, w Workload) error {
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
func WorkloadResultPush(ctx context.Context, db *mongo.Database, id schema.ID, result Result) error {
	col := db.Collection(WorkloadCollection)
	var filter WorkloadFilter
	filter = filter.WithID(id)

	// we don't care if we couldn't delete old result.
	// in case it never existed, or the array is nil.
	col.UpdateOne(ctx, filter, bson.M{
		"$pull": bson.M{
			"results": bson.M{
				"workload_id": result.WorkloadId,
				"node_id":     result.NodeId,
			},
		},
	})

	_, err := col.UpdateOne(ctx, filter, bson.D{
		{
			Key: "$push",
			Value: bson.M{
				"results": result,
			},
		},
	})

	return err
}

// Workload returns workload
func (w *WorkloaderType) Workload() Workload {
	return Workload{
		ReservationWorkload: generated.ReservationWorkload{
			WorkloadId: fmt.Sprintf("%d-%d", w.GetID(), w.WorkloadID()),
			PoolID:     w.GetPoolID(),
			User:       fmt.Sprint(w.GetCustomerTid()),
			Type:       w.GetWorkloadType(),
			Duration:   math.MaxInt64,
			Created:    w.GetEpoch(),
			ToDelete:   w.GetNextAction() == Delete || w.GetNextAction() == Deleted,
		},
		NodeID: w.GetNodeID(),
	}
}
