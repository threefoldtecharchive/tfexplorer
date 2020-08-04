package types

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

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
	// ReservationCollection db collection name
	ReservationCollection = "reservation"
	queueCollection       = "workqueue"
)

const (
	// Create action
	Create = generated.NextActionCreate
	// Sign action
	Sign = generated.NextActionSign
	// Pay action
	Pay = generated.NextActionPay
	// Deploy action
	Deploy = generated.NextActionDeploy
	// Delete action
	Delete = generated.NextActionDelete
	// Invalid action
	Invalid = generated.NextActionInvalid
	// Deleted action
	Deleted = generated.NextActionDeleted
)

// ApplyQueryFilter parese the query string
func ApplyQueryFilter(r *http.Request, filter ReservationFilter) (ReservationFilter, error) {
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

// ReservationFilter type
type ReservationFilter bson.D

// WithID filter reservation with ID
func (f ReservationFilter) WithID(id schema.ID) ReservationFilter {
	return append(f, bson.E{Key: "_id", Value: id})
}

// WithIDGE return find reservations with
func (f ReservationFilter) WithIDGE(id schema.ID) ReservationFilter {
	return append(f, bson.E{
		Key: "_id", Value: bson.M{"$gte": id},
	})
}

// WithNextAction filter reservations with next action
func (f ReservationFilter) WithNextAction(action generated.NextActionEnum) ReservationFilter {
	return append(f, bson.E{
		Key: "next_action", Value: action,
	})
}

// WithCustomerID filter reservation on customer
func (f ReservationFilter) WithCustomerID(customerID int64) ReservationFilter {
	return append(f, bson.E{
		Key: "customer_tid", Value: customerID,
	})

}

// WithNodeID searsch reservations with NodeID
func (f ReservationFilter) WithNodeID(id string) ReservationFilter {
	//data_reservation.{containers, volumes, zdbs, networks, kubernetes}.node_id
	// we need to search ALL types for any reservation that has the node ID
	or := []bson.M{}

	for _, typ := range []string{"containers", "volumes", "zdbs", "kubernetes", "proxies", "reverse_proxies", "subdomains", "domain_delegates", "gateway4to6", "capacity_pool"} {
		key := fmt.Sprintf("data_reservation.%s.node_id", typ)
		or = append(or, bson.M{key: id})
	}

	// network workload is special because node id is set on the network_resources.
	or = append(or, bson.M{"data_reservation.networks.network_resources.node_id": id})

	// we find any reservation that has this node ID set.
	return append(f, bson.E{Key: "$or", Value: or})
}

// Or returns filter that reads as (f or o)
func (f ReservationFilter) Or(o ReservationFilter) ReservationFilter {
	return ReservationFilter{
		bson.E{
			Key:   "$or",
			Value: bson.A{f, o},
		},
	}
}

// Get gets single reservation that matches the filter
func (f ReservationFilter) Get(ctx context.Context, db *mongo.Database) (reservation Reservation, err error) {
	if f == nil {
		f = ReservationFilter{}
	}

	result := db.Collection(ReservationCollection).FindOne(ctx, f)
	if err = result.Err(); err != nil {
		return
	}

	err = result.Decode(&reservation)
	return
}

// Find all users that matches filter
func (f ReservationFilter) Find(ctx context.Context, db *mongo.Database, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	if f == nil {
		f = ReservationFilter{}
	}
	return db.Collection(ReservationCollection).Find(ctx, f, opts...)
}

// Count number of documents matching
func (f ReservationFilter) Count(ctx context.Context, db *mongo.Database) (int64, error) {
	col := db.Collection(ReservationCollection)
	if f == nil {
		f = ReservationFilter{}
	}

	return col.CountDocuments(ctx, f)
}

// Reservation is a wrapper around generated type
type Reservation generated.Reservation

// Validate that the reservation is valid
func (r *Reservation) Validate() error {
	if r.CustomerTid == 0 {
		return fmt.Errorf("customer_tid is required")
	}

	if len(r.CustomerSignature) == 0 {
		return fmt.Errorf("customer_signature is required")
	}

	var data generated.ReservationData

	if err := json.Unmarshal([]byte(r.Json), &data); err != nil {
		return errors.Wrap(err, "invalid json data on reservation")
	}

	if !reflect.DeepEqual(r.DataReservation, data) {
		return fmt.Errorf("json data does not match the reservation data")
	}

	if len(r.Metadata) > 1024 {
		return fmt.Errorf("metadata can not be bigger than 1024 bytes")
	}

	totalWl := len(r.DataReservation.Containers) +
		len(r.DataReservation.Networks) +
		len(r.DataReservation.NetworkResources) +
		len(r.DataReservation.Volumes) +
		len(r.DataReservation.Zdbs) +
		len(r.DataReservation.Proxies) +
		len(r.DataReservation.ReverseProxy) +
		len(r.DataReservation.Subdomains) +
		len(r.DataReservation.DomainDelegates) +
		len(r.DataReservation.Gateway4To6s)

	// all workloads are supposed to implement this interface
	type workloader interface{ WorkloadID() int64 }

	ids := make(map[int64]struct{}, totalWl)
	workloaders := make([]workloader, 0, totalWl)

	// seems go doesn't allow : workloaders=append(workloaders, r.DataReservation.Containers)
	// so we have to loop
	for _, w := range r.DataReservation.Containers {
		if w.Capacity.DiskType != generated.DiskTypeSSD {
			return errors.New("Container disktype is not valid, it should be SSD")
		}
		workloaders = append(workloaders, &w)
	}
	for _, w := range r.DataReservation.Volumes {
		if w.Type != generated.VolumeTypeSSD {
			return errors.New("Volume disktype is not valid, it should be SSD")
		}
		workloaders = append(workloaders, &w)
	}
	for _, w := range r.DataReservation.Zdbs {
		workloaders = append(workloaders, &w)
	}
	for _, w := range r.DataReservation.Kubernetes {
		workloaders = append(workloaders, &w)
	}
	for _, w := range r.DataReservation.Proxies {
		workloaders = append(workloaders, &w)
	}
	for _, w := range r.DataReservation.ReverseProxy {
		workloaders = append(workloaders, &w)
	}
	for _, w := range r.DataReservation.Subdomains {
		workloaders = append(workloaders, &w)
	}
	for _, w := range r.DataReservation.DomainDelegates {
		workloaders = append(workloaders, &w)
	}
	for _, w := range r.DataReservation.Gateway4To6s {
		workloaders = append(workloaders, &w)
	}

	for _, w := range workloaders {
		if _, ok := ids[w.WorkloadID()]; ok {
			return fmt.Errorf("conflicting workload ID '%d'", w.WorkloadID())
		}
		ids[w.WorkloadID()] = struct{}{}
	}

	return nil
}

// Verify signature against Reserveration.JSON
// pk is the public key used as verification key in hex encoded format
// the signature is the signature to verify (in raw binary format)
func (r *Reservation) Verify(pk string, sig []byte) error {
	key, err := crypto.KeyFromHex(pk)
	if err != nil {
		return errors.Wrap(err, "invalid verification key")
	}

	return crypto.Verify(key, []byte(r.Json), sig)
}

// SignatureVerify is similar to Verify but the verification is done
// against `str(Reservation.ID) + Reservation.JSON`
func (r *Reservation) SignatureVerify(pk string, sig []byte) error {
	key, err := crypto.KeyFromHex(pk)
	if err != nil {
		return errors.Wrap(err, "invalid verification key")
	}

	var buf bytes.Buffer
	if _, err := buf.WriteString(fmt.Sprint(int64(r.ID))); err != nil {
		return errors.Wrap(err, "failed to write id to buffer")
	}

	if _, err := buf.WriteString(r.Json); err != nil {
		return errors.Wrap(err, "failed to write json to buffer")
	}

	return crypto.Verify(key, buf.Bytes(), sig)
}

// Expired checks if this reservation has expired
func (r *Reservation) Expired() bool {
	return time.Until(r.DataReservation.ExpirationReservation.Time) <= 0
}

// IsAny checks if the reservation status is any of the given status
func (r *Reservation) IsAny(status ...generated.NextActionEnum) bool {
	for _, s := range status {
		if r.NextAction == s {
			return true
		}
	}

	return false
}

//ResultOf return result of a workload ID
func (r *Reservation) ResultOf(id string) *Result {
	for _, result := range r.Results {
		if result.WorkloadId == id {
			r := Result(result)
			return &r
		}
	}

	return nil
}

// AllDeleted checks of all workloads has been marked
func (r *Reservation) AllDeleted() bool {
	// check if all workloads have been deleted.
	for _, wl := range r.Workloads("") {
		result := r.ResultOf(wl.UniqueWorkloadID())
		if result == nil ||
			result.State != generated.ResultStateDeleted {
			return false
		}
	}

	return true
}

// Workloads returns all reservation workloads (filter by nodeID)
// if nodeID is empty, return all workloads
func (r *Reservation) Workloads(nodeID string) []WorkloaderType {

	data := &r.DataReservation

	newWrkl := func(w workloads.Workloader, id schema.ID, user int64, na workloads.NextActionEnum, epoch schema.Date, meta string, provisionRequest workloads.SigningRequest, deleteRequest workloads.SigningRequest, result *Result) WorkloaderType {
		workload := WorkloaderType{Workloader: w}
		workload.SetCustomerTid(user)
		workload.SetNextAction(na)
		workload.SetID(id)
		workload.SetEpoch(epoch)
		workload.SetMetadata(meta)
		if result != nil {
			workload.SetResult(workloads.Result(*result))
		}
		workload.SetSigningRequestProvision(provisionRequest)
		workload.SetSigningRequestDelete(deleteRequest)

		return workload
	}

	var wrklds []WorkloaderType
	for i := range data.Containers {
		wl := data.Containers[i]
		if len(nodeID) > 0 && wl.NodeId != nodeID {
			continue
		}
		uwid := fmt.Sprintf("%d-%d", r.ID, wl.WorkloadId)
		wl.WorkloadType = generated.WorkloadTypeContainer
		wrklds = append(wrklds, newWrkl(
			&wl,
			r.ID,
			r.CustomerTid,
			r.NextAction,
			r.Epoch,
			r.Metadata,
			r.DataReservation.SigningRequestProvision,
			r.DataReservation.SigningRequestDelete,
			r.ResultOf(uwid)))
	}

	for i := range data.Volumes {
		wl := data.Volumes[i]
		if len(nodeID) > 0 && wl.NodeId != nodeID {
			continue
		}
		uwid := fmt.Sprintf("%d-%d", r.ID, wl.WorkloadID())
		wl.WorkloadType = generated.WorkloadTypeVolume
		wrklds = append(wrklds, newWrkl(
			&wl,
			r.ID,
			r.CustomerTid,
			r.NextAction,
			r.Epoch,
			r.Metadata,
			r.DataReservation.SigningRequestProvision,
			r.DataReservation.SigningRequestDelete,
			r.ResultOf(uwid)))
	}
	for i := range data.Zdbs {
		wl := data.Zdbs[i]
		if len(nodeID) > 0 && wl.NodeId != nodeID {
			continue
		}
		uwid := fmt.Sprintf("%d-%d", r.ID, wl.WorkloadID())
		wl.WorkloadType = generated.WorkloadTypeZDB
		wrklds = append(wrklds, newWrkl(
			&wl,
			r.ID,
			r.CustomerTid,
			r.NextAction,
			r.Epoch,
			r.Metadata,
			r.DataReservation.SigningRequestProvision,
			r.DataReservation.SigningRequestDelete,
			r.ResultOf(uwid)))
	}
	for i := range data.Kubernetes {
		wl := data.Kubernetes[i]
		if len(nodeID) > 0 && wl.NodeId != nodeID {
			continue
		}
		uwid := fmt.Sprintf("%d-%d", r.ID, wl.WorkloadID())
		wl.WorkloadType = generated.WorkloadTypeKubernetes
		wrklds = append(wrklds, newWrkl(
			&wl,
			r.ID,
			r.CustomerTid,
			r.NextAction,
			r.Epoch,
			r.Metadata,
			r.DataReservation.SigningRequestProvision,
			r.DataReservation.SigningRequestDelete,
			r.ResultOf(uwid)))
	}
	for i := range data.Proxies {
		wl := data.Proxies[i]
		if len(nodeID) > 0 && wl.NodeId != nodeID {
			continue
		}
		uwid := fmt.Sprintf("%d-%d", r.ID, wl.WorkloadID())
		wl.WorkloadType = generated.WorkloadTypeProxy
		wrklds = append(wrklds, newWrkl(
			&wl,
			r.ID,
			r.CustomerTid,
			r.NextAction,
			r.Epoch,
			r.Metadata,
			r.DataReservation.SigningRequestProvision,
			r.DataReservation.SigningRequestDelete,
			r.ResultOf(uwid)))
	}
	for i := range data.ReverseProxy {
		wl := data.ReverseProxy[i]
		if len(nodeID) > 0 && wl.NodeId != nodeID {
			continue
		}
		uwid := fmt.Sprintf("%d-%d", r.ID, wl.WorkloadID())
		wl.WorkloadType = generated.WorkloadTypeReverseProxy
		wrklds = append(wrklds, newWrkl(
			&wl,
			r.ID,
			r.CustomerTid,
			r.NextAction,
			r.Epoch,
			r.Metadata,
			r.DataReservation.SigningRequestProvision,
			r.DataReservation.SigningRequestDelete,
			r.ResultOf(uwid)))
	}
	for i := range data.Subdomains {
		wl := data.Subdomains[i]
		if len(nodeID) > 0 && wl.NodeId != nodeID {
			continue
		}
		uwid := fmt.Sprintf("%d-%d", r.ID, wl.WorkloadID())
		wl.WorkloadType = generated.WorkloadTypeSubDomain
		wrklds = append(wrklds, newWrkl(
			&wl,
			r.ID,
			r.CustomerTid,
			r.NextAction,
			r.Epoch,
			r.Metadata,
			r.DataReservation.SigningRequestProvision,
			r.DataReservation.SigningRequestDelete,
			r.ResultOf(uwid)))
	}
	for i := range data.DomainDelegates {
		wl := data.DomainDelegates[i]
		if len(nodeID) > 0 && wl.NodeId != nodeID {
			continue
		}
		uwid := fmt.Sprintf("%d-%d", r.ID, wl.WorkloadID())
		wl.WorkloadType = generated.WorkloadTypeDomainDelegate
		wrklds = append(wrklds, newWrkl(
			&wl,
			r.ID,
			r.CustomerTid,
			r.NextAction,
			r.Epoch,
			r.Metadata,
			r.DataReservation.SigningRequestProvision,
			r.DataReservation.SigningRequestDelete,
			r.ResultOf(uwid)))
	}
	for i := range data.Gateway4To6s {
		wl := data.Gateway4To6s[i]
		if len(nodeID) > 0 && wl.NodeId != nodeID {
			continue
		}
		uwid := fmt.Sprintf("%d-%d", r.ID, wl.WorkloadID())
		wl.WorkloadType = generated.WorkloadTypeGateway4To6
		wrklds = append(wrklds, newWrkl(
			&wl,
			r.ID,
			r.CustomerTid,
			r.NextAction,
			r.Epoch,
			r.Metadata,
			r.DataReservation.SigningRequestProvision,
			r.DataReservation.SigningRequestDelete,
			r.ResultOf(uwid)))
	}
	for i := range data.Networks {
		wl := data.Networks[i]
		networkResources := wl.ToNetworkResources()
		for i := range networkResources {
			nr := networkResources[i]
			if len(nodeID) > 0 && nr.NodeId != nodeID {
				continue
			}

			uwid := fmt.Sprintf("%d-%d", r.ID, wl.WorkloadID())
			nr.WorkloadType = generated.WorkloadTypeNetworkResource
			wrklds = append(wrklds, newWrkl(
				&nr,
				r.ID,
				r.CustomerTid,
				r.NextAction,
				r.Epoch,
				r.Metadata,
				r.DataReservation.SigningRequestProvision,
				r.DataReservation.SigningRequestDelete,
				r.ResultOf(uwid)))
		}
	}
	for i := range data.NetworkResources {
		wl := data.NetworkResources[i]
		if len(nodeID) > 0 && wl.NodeId != nodeID {
			continue
		}
		uwid := fmt.Sprintf("%d-%d", r.ID, wl.WorkloadID())
		wl.WorkloadType = generated.WorkloadTypeNetworkResource
		wrklds = append(wrklds, newWrkl(
			&wl,
			r.ID,
			r.CustomerTid,
			r.NextAction,
			r.Epoch,
			r.Metadata,
			r.DataReservation.SigningRequestProvision,
			r.DataReservation.SigningRequestDelete,
			r.ResultOf(uwid)))
	}

	return wrklds
}

// IsSuccessfullyDeployed check if all the workloads defined in the reservation
// have sent a positive result
func (r *Reservation) IsSuccessfullyDeployed() bool {
	succeeded := false
	if len(r.Results) >= len(r.Workloads("")) {
		succeeded = true
		for _, result := range r.Results {
			if result.State != generated.ResultStateOK {
				succeeded = false
				break
			}
		}
	}
	return succeeded
}

// NodeIDs used by this reservation
func (r *Reservation) NodeIDs() []string {
	ids := make(map[string]struct{})
	for _, w := range r.DataReservation.Containers {
		ids[w.NodeId] = struct{}{}
	}

	for _, w := range r.DataReservation.Networks {
		for _, nr := range w.NetworkResources {
			ids[nr.NodeId] = struct{}{}
		}
	}

	for _, w := range r.DataReservation.NetworkResources {
		ids[w.NodeId] = struct{}{}
	}

	for _, w := range r.DataReservation.Zdbs {
		ids[w.NodeId] = struct{}{}
	}

	for _, w := range r.DataReservation.Volumes {
		ids[w.NodeId] = struct{}{}
	}

	for _, w := range r.DataReservation.Kubernetes {
		ids[w.NodeId] = struct{}{}
	}

	nodeIDs := make([]string, 0, len(ids))
	for nid := range ids {
		nodeIDs = append(nodeIDs, nid)
	}
	return nodeIDs
}

// GatewayIDs return a list of all the gateway IDs used in this reservation
func (r *Reservation) GatewayIDs() []string {
	ids := make(map[string]struct{})

	for _, p := range r.DataReservation.Proxies {
		ids[p.NodeId] = struct{}{}
	}

	for _, p := range r.DataReservation.ReverseProxy {
		ids[p.NodeId] = struct{}{}
	}

	for _, p := range r.DataReservation.Subdomains {
		ids[p.NodeId] = struct{}{}
	}

	for _, p := range r.DataReservation.DomainDelegates {
		ids[p.NodeId] = struct{}{}
	}

	for _, p := range r.DataReservation.Gateway4To6s {
		ids[p.NodeId] = struct{}{}
	}

	gwIDs := make([]string, 0, len(ids))
	for nid := range ids {
		gwIDs = append(gwIDs, nid)
	}
	return gwIDs
}

// ReservationCreate save new reservation to database.
// NOTE: use reservations only that are returned from calling Pipeline.Next()
// no validation is done here, this is just a CRUD operation
func ReservationCreate(ctx context.Context, db *mongo.Database, r Reservation) (schema.ID, error) {
	id := models.MustID(ctx, db, ReservationCollection)
	r.ID = id

	_, err := db.Collection(ReservationCollection).InsertOne(ctx, r)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// ReservationLastID get the current last ID number in the reservations collection
func ReservationLastID(ctx context.Context, db *mongo.Database) (schema.ID, error) {
	return models.LastID(ctx, db, ReservationCollection)
}

// ReservationSetNextAction update the reservation next action in db
func ReservationSetNextAction(ctx context.Context, db *mongo.Database, id schema.ID, action generated.NextActionEnum) error {
	var filter ReservationFilter
	filter = filter.WithID(id)

	col := db.Collection(ReservationCollection)
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

// ReservationToDeploy marks a reservation to deploy and schedule the workloads for the nodes
// it's a short cut to SetNextAction then PushWorkloads
func ReservationToDeploy(ctx context.Context, db *mongo.Database, reservation *Reservation) error {
	// update reservation
	if err := ReservationSetNextAction(ctx, db, reservation.ID, Deploy); err != nil {
		return errors.Wrap(err, "failed to set reservation to DEPLOY state")
	}

	//queue for processing
	if err := WorkloadPush(ctx, db, reservation.Workloads("")...); err != nil {
		return errors.Wrap(err, "failed to schedule reservation for deploying")
	}

	return nil
}

// SignatureMode type
type SignatureMode string

const (
	// SignatureProvision mode
	SignatureProvision SignatureMode = "signatures_provision"
	// SignatureDelete mode
	SignatureDelete SignatureMode = "signatures_delete"
)

//ReservationPushSignature push signature to reservation
func ReservationPushSignature(ctx context.Context, db *mongo.Database, id schema.ID, mode SignatureMode, signature generated.SigningSignature) error {
	// this function just push the signature to the reservation array
	// there are not other checks involved here. So before calling this function
	// we need to ensure the signature has the rights to be pushed
	var filter ReservationFilter
	filter = filter.WithID(id)
	col := db.Collection(ReservationCollection)
	_, err := col.UpdateOne(ctx, filter, bson.M{
		"$push": bson.M{
			string(mode): signature,
		},
	})

	return err
}

// Workload is a wrapper around generated TfgridWorkloadsReservationWorkload1 type
type Workload struct {
	generated.ReservationWorkload `bson:",inline"`
	NodeID                        string `json:"node_id" bson:"node_id"`
}

// QueueFilter for workloads in temporary queue
type QueueFilter bson.D

// WithNodeID search queue with node-id
func (f QueueFilter) WithNodeID(nodeID string) QueueFilter {
	return append(f, bson.E{Key: "node_id", Value: nodeID})
}

// Find runs the filter, and return a cursor
func (f QueueFilter) Find(ctx context.Context, db *mongo.Database, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	col := db.Collection(queueCollection)
	return col.Find(ctx, f, opts...)
}

// WorkloadPush pushes a workload to the queue
func WorkloadPush(ctx context.Context, db *mongo.Database, w ...WorkloaderType) error {
	col := db.Collection(queueCollection)

	for _, wl := range w {
		_, err := col.UpdateOne(ctx, bson.M{"_id": wl.GetID()}, bson.M{"$set": wl}, options.Update().SetUpsert(true))
		if err != nil {
			return errors.Wrap(err, "could not upsert workload")
		}
	}

	return nil
}

// WorkloadPop removes workload from queue
func WorkloadPop(ctx context.Context, db *mongo.Database, id schema.ID) error {
	col := db.Collection(queueCollection)
	_, err := col.DeleteOne(ctx, bson.M{"_id": id})

	return err
}

// Result is a wrapper around TfgridWorkloadsReservationResult1 type
type Result generated.Result

// Verify that the signature matches the result data
func (r *Result) Verify(pk string) error {
	return nil
	// sig, err := hex.DecodeString(r.Signature)
	// if err != nil {
	// 	return errors.Wrap(err, "invalid signature expecting hex encoded")
	// }

	// key, err := crypto.KeyFromID(pkg.StrIdentifier(pk))
	// if err != nil {
	// 	return errors.Wrap(err, "invalid verification key")
	// }

	// bytes, err := r.encode()
	// if err != nil {
	// 	return err
	// }

	// return crypto.Verify(key, bytes, sig)
}

// ResultPush pushes result to a reservation result array.
// NOTE: this is just a crud operation, no validation is done here
func ResultPush(ctx context.Context, db *mongo.Database, id schema.ID, result Result) error {
	col := db.Collection(ReservationCollection)
	var filter ReservationFilter
	filter = filter.WithID(id)

	// we don't care if we couldn't delete old result.
	// in case it never existed, or the array is nil.
	_, err := col.UpdateOne(ctx, filter, bson.D{
		{
			Key: "$set",
			Value: bson.M{
				"result": result,
			},
		},
	})

	return err
}
