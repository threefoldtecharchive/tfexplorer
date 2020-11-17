package types

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/models"
	"github.com/threefoldtech/tfexplorer/schema"
	"github.com/threefoldtech/zos/pkg/crypto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type (
	// Reservation is a reservation type to create a new capacity pool, or
	// to add capacity to an existing pool. Ownership of a new pool is tied to
	// the person who signs the request to set up the pool. Once a pool is
	// set up, anyone can top it up with additional capacity. Even though there
	// is no restriction of who can add capacity to the pool, only the owner
	// can assign workloads to it.
	//
	// This type is based on the `generated.Reservation` type. The main reason
	// for this being a separate type, is that the aforementioned type is actually
	// not flexible, and strongly tied to regular workloads. One solution would
	// be to make a pool a type of workload, but this would be a serious hack.
	//
	// Furthermore, note that some fields have been stripped. Reason is, that a
	// capacity pool is only meant to serve as an abstract concept, internal to
	// the explorer, and later the farmer threebot. As such, there are no dedicated
	// signature fields. Other workload specific info is also stripped. Note that
	// the way of signing the reservation is kept. While this method is questionable
	// to say the least, it does mean that we will have a much easier time if we
	// decide to merge the 2 reservation types in the future, which we should still
	// do.
	Reservation struct {
		ID                schema.ID       `bson:"_id" json:"id"`
		JSON              string          `bson:"json" json:"json"`
		DataReservation   ReservationData `bson:"data_reservation" json:"data_reservation"`
		CustomerTid       int64           `bson:"customer_tid" json:"customer_tid"`
		CustomerSignature string          `bson:"customer_signature" json:"customer_signature"`
	}

	// ReservationData is the actual data sent in a capacity pool reservation. If
	// PoolID is a non-zero value, this reservation will add the requested capacity
	// to the existing pool with the given ID.
	//
	// Although CU and SU values for workloads can be (and likely will be) floating
	// points, we only allow purchasing full units. Since such a unit is actually
	// very small, this is not a problem for over purchasing, and it simplifies
	// some stuff on our end.
	ReservationData struct {
		PoolID     int64    `bson:"pool_id" json:"pool_id"`
		CUs        uint64   `bson:"cus" json:"cus"`
		SUs        uint64   `bson:"sus" json:"sus"`
		IPv4Us     uint64   `bson:"ipv4u" json:"ipv4u"`
		NodeIDs    []string `bson:"node_ids" json:"node_ids"`
		Currencies []string `bson:"currencies" json:"currencies"`
	}
)

const (
	// CapacityReservationCollection db collection name
	CapacityReservationCollection = "capacity-reservations"
)

// Validate the reservation
func (pr *Reservation) Validate() error {

	if pr.CustomerTid == 0 {
		return errors.New("customer_tid is required")
	}

	if len(pr.CustomerSignature) == 0 {
		return errors.New("customer_signature is required")
	}

	if len(pr.DataReservation.NodeIDs) == 0 {
		return errors.New("pool must be applicable to at least 1 node")
	}

	var data ReservationData

	if err := json.Unmarshal([]byte(pr.JSON), &data); err != nil {
		return errors.Wrap(err, "invalid json data on reservation")
	}

	if !reflect.DeepEqual(pr.DataReservation, data) {
		return fmt.Errorf("json data does not match the reservation data")
	}

	return nil
}

// Verify the provided signature against the reservation JSON, with the provided
// key. The key is the public key of the user, as a hex string
func (pr *Reservation) Verify(pk string) error {
	signature, err := hex.DecodeString(pr.CustomerSignature)
	if err != nil {
		return errors.Wrap(err, "invalid signature format, expecting hex encoded string")
	}
	key, err := crypto.KeyFromHex(pk)
	if err != nil {
		return errors.Wrap(err, "invalid verification key")
	}

	return crypto.Verify(key, []byte(pr.JSON), signature)
}

// CapacityReservationCreate saves a new capacity reservation to the database
func CapacityReservationCreate(ctx context.Context, db *mongo.Database, reservation Reservation) (Reservation, error) {
	id := models.MustID(ctx, db, CapacityReservationCollection)
	reservation.ID = id

	_, err := db.Collection(CapacityReservationCollection).InsertOne(ctx, reservation)
	if err != nil {
		return reservation, err
	}

	return reservation, nil
}

// CapacityReservationGet loads a capacity reservation with the given id
func CapacityReservationGet(ctx context.Context, db *mongo.Database, id schema.ID) (Reservation, error) {
	var reservation Reservation
	res := db.Collection(CapacityReservationCollection).FindOne(ctx, bson.M{"_id": id})
	if err := res.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return reservation, ErrPoolNotFound
		}
		return reservation, errors.Wrap(err, "could not load reservation")
	}
	err := res.Decode(&reservation)
	return reservation, err
}
