package types

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/models"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
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
		filter = filter.WithCustomerID(customerid)
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
func (f WorkloadFilter) WithCustomerID(customerID int64) WorkloadFilter {
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
func (f WorkloadFilter) Get(ctx context.Context, db *mongo.Database) (WorkloaderType, error) {
	if f == nil {
		f = WorkloadFilter{}
	}
	var w WorkloaderType

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
func (f WorkloadFilter) Find(ctx context.Context, db *mongo.Database, opts ...*options.FindOptions) ([]WorkloaderType, error) {
	if f == nil {
		f = WorkloadFilter{}
	}

	cursor, err := db.Collection(WorkloadCollection).Find(ctx, f, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get workload cursor")
	}

	ws := []WorkloaderType{}

	for cursor.Next(ctx) {
		var w WorkloaderType
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

// WorkloaderType is a wrapper struct around the Workloader interface
type WorkloaderType struct {
	generated.Workloader
}

// MarshalBSON implements bson.Marshaller
func (w WorkloaderType) MarshalBSON() ([]byte, error) {
	return bson.Marshal(w.Workloader)
}

// UnmarshalBSON implements bson.Unmarshaller
func (w *WorkloaderType) UnmarshalBSON(buf []byte) error {
	workload, err := workloads.UnmarshalBSON(buf)
	if err != nil {
		return err
	}

	if w == nil {
		w = &WorkloaderType{}
	}

	*w = WorkloaderType{Workloader: workload}

	return nil
}

// MarshalJSON implements JSON.Marshaller
func (w WorkloaderType) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.Workloader)
}

// UnmarshalJSON implements JSON.Unmarshaller
func (w *WorkloaderType) UnmarshalJSON(buf []byte) error {
	workload, err := workloads.UnmarshalJSON(buf)
	if err != nil {
		return err
	}

	if w == nil {
		w = &WorkloaderType{}
	}

	*w = WorkloaderType{Workloader: workload}

	return nil
}

// Verify signature against Workload.JSON
// pk is the public key used as verification key in hex encoded format
// the signature is the signature to verify (in raw binary format)
func (w *WorkloaderType) Verify(pk string, sig []byte) error {
	key, err := crypto.KeyFromHex(pk)
	if err != nil {
		return errors.Wrap(err, "invalid verification key")
	}

	b, err := w.SignatureChallenge()
	if err != nil {
		return err
	}

	msg := sha256.Sum256(b)

	return crypto.Verify(key, msg[:], sig)
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

// SignatureDeleteRequestVerify verify the signature from a signature request
// this is used for workload delete
// the signature is created from the workload siging challenge + "delete" + customer tid
func (w *WorkloaderType) SignatureDeleteRequestVerify(pk string, sig generated.SigningSignature) error {
	key, err := crypto.KeyFromHex(pk)
	if err != nil {
		return errors.Wrap(err, "invalid verification key")
	}

	b, err := w.SignatureChallenge()
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)
	if _, err := buf.WriteString("delete"); err != nil {
		return err
	}
	if _, err := buf.WriteString(fmt.Sprintf("%d", sig.Tid)); err != nil {
		return err
	}

	msg := sha256.Sum256(buf.Bytes())
	signature, err := hex.DecodeString(sig.Signature)
	if err != nil {
		return err
	}

	return crypto.Verify(key, msg[:], signature)
}

// SignatureProvisionRequestVerify verify the signature from a signature request
// this is used for provision
// the signature is created from the workload siging challenge + "provision" + customer tid
func (w *WorkloaderType) SignatureProvisionRequestVerify(pk string, sig generated.SigningSignature) error {
	key, err := crypto.KeyFromHex(pk)
	if err != nil {
		return errors.Wrap(err, "invalid verification key")
	}

	b, err := w.SignatureChallenge()
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)
	if _, err := buf.WriteString("provision"); err != nil {
		return err
	}
	if _, err := buf.WriteString(fmt.Sprintf("%d", sig.Tid)); err != nil {
		return err
	}

	msg := sha256.Sum256(buf.Bytes())
	signature, err := hex.DecodeString(sig.Signature)
	if err != nil {
		return err
	}

	return crypto.Verify(key, msg[:], signature)
}

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
	return w.GetResult().State == generated.ResultStateDeleted
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

	// ensure the signers array are never nill cause it would cause issue
	// in mongo when trying to push signature later on
	reqDel := w.GetSigningRequestDelete()
	if reqDel.Signers == nil {
		reqDel.Signers = make([]int64, 0)
		w.SetSigningRequestDelete(reqDel)
	}

	reqPro := w.GetSigningRequestProvision()
	if reqPro.Signers == nil {
		reqDel.Signers = make([]int64, 0)
		w.SetSigningRequestProvision(reqPro)
	}

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
	if err := WorkloadTypePush(ctx, db, w); err != nil {
		return errors.Wrap(err, "failed to schedule workload for deploying")
	}

	return nil
}

//WorkloadPushSignature push signature to workload
func WorkloadPushSignature(ctx context.Context, db *mongo.Database, id schema.ID, mode SignatureMode, signature generated.SigningSignature) error {
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
func WorkloadTypePush(ctx context.Context, db *mongo.Database, w WorkloaderType) error {
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

	_, err := col.UpdateOne(ctx, filter,
		bson.M{
			"$set": bson.M{
				"result": result,
			},
		},
	)

	return err
}

// Validate that the reservation is valid
func (w *WorkloaderType) Validate() error {
	if w.GetCustomerTid() == 0 {
		return fmt.Errorf("customer_tid is required")
	}

	if len(w.GetCustomerSignature()) == 0 {
		return fmt.Errorf("customer_signature is required")
	}

	if len(w.GetMetadata()) > 1024 {
		return fmt.Errorf("metadata can not be bigger than 1024 bytes")
	}

	if w.GetPoolID() == 0 {
		return errors.New("pool is required")
	}

	if w.GetReference() != "" {
		return errors.New("reference is illegal for new workloads")
	}

	return nil
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
			Content:    w,
		},
		NodeID: w.GetNodeID(),
	}
}
