package types

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/stellar/go/xdr"

	"github.com/threefoldtech/tfexplorer/pkg/stellar"
	"github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// EscrowCollection db collection name
	EscrowCollection = "escrow"
	// CapacityEscrowCollection db collection for escrow info related to capacity info
	CapacityEscrowCollection = "capacity-escrow"
)

var (
	// ErrEscrowExists is returned when trying to save escrow information for a
	// reservation that already has escrow information
	ErrEscrowExists = errors.New("escrow(s) for reservation already exists")
	// ErrEscrowNotFound is returned if escrow information is not found
	ErrEscrowNotFound = errors.New("escrow information not found")
)

type (
	// ReservationPaymentInformation stores the reservation payment information
	ReservationPaymentInformation struct {
		ReservationID schema.ID      `bson:"_id"`
		Address       string         `bson:"address"`
		Expiration    schema.Date    `bson:"expiration"`
		Asset         stellar.Asset  `bson:"asset"`
		Infos         []EscrowDetail `bson:"infos"`
		// Paid indicates the reservation escrows have been fully funded, and
		// the reservation has been moved from the "PAY" state to the "DEPLOY"
		// state
		Paid bool `bson:"paid"`
		// Released indicates the reservation has been fully deployed, and
		// that an attempt was made to pay the farmers. If this flag is set, it is
		// still possible that there are funds on an escrow related to this transaction
		// either because someone funded it after the reservation was already
		// deployed , because there was an error paying the farmers
		// or because there was an error refunding any overpaid amount.
		Released bool `bson:"released"`
		// Canceled indicates that the reservation was canceled, either by the
		// user, or because a workload deployment failed, which resulted in the
		// entire reservation being canceled. As a result, an attempt was made
		// to refund the client. It is possible for this to have failed.
		Canceled bool `bson:"canceled"`
	}

	// CapacityReservationPaymentInformation stores the reservation payment information
	CapacityReservationPaymentInformation struct {
		ReservationID schema.ID     `bson:"_id"`
		FarmerID      schema.ID     `bson:"farmer_id"`
		Address       string        `bson:"address"`
		Expiration    schema.Date   `bson:"expiration"`
		Asset         stellar.Asset `bson:"asset"`
		Amount        xdr.Int64     `bson:"amount"`
		// Paid indicates the capacity reservation escrows have been fully funded,
		// resulting in the new funds being allocated into the pool (creating
		// the pool in case it did not exist yet)
		Paid bool `bson:"paid"`
		// Released means we tried to pay the farmer
		Released bool `bson:"released"`
		// Canceled means the escrow got canceled, i.e. client refunded. This can
		// only happen in case the reservation expires.
		Canceled bool `bson:"canceled"`
	}

	// EscrowDetail hold the details of an escrow address
	EscrowDetail struct {
		FarmerID    schema.ID `bson:"farmer_id" json:"farmer_id"`
		TotalAmount xdr.Int64 `bson:"total_amount" json:"total_amount"`
	}

	// CustomerEscrowInformation is the escrow information which will get exposed
	// to the customer once he creates a reservation
	CustomerEscrowInformation struct {
		Address string         `json:"address"`
		Asset   stellar.Asset  `json:"asset"`
		Details []EscrowDetail `json:"details"`
	}

	// CustomerCapacityEscrowInformation is the escrow information which will get exposed
	// to the customer once he creates a reservation for capacity
	CustomerCapacityEscrowInformation struct {
		Address string        `json:"address"`
		Asset   stellar.Asset `json:"asset"`
		Amount  xdr.Int64     `json:"amount"`
	}

	// CapacityReservationInfo is information to manipulate a capacity pool once
	// the escrow is fulfilled
	CapacityReservationInfo struct {
		// ID of the pool
		ID schema.ID `bson:"id"`
		// CUs to add
		CUs uint64 `bson:"cus"`
		// SUs to add
		SUs uint64 `bson:"sus"`
	}
)

// ReservationPaymentInfoCreate creates the reservation payment information
func ReservationPaymentInfoCreate(ctx context.Context, db *mongo.Database, reservationPaymentInfo ReservationPaymentInformation) error {
	col := db.Collection(EscrowCollection)
	_, err := col.InsertOne(ctx, reservationPaymentInfo)
	if err != nil {
		if merr, ok := err.(mongo.WriteException); ok {
			errCode := merr.WriteErrors[0].Code
			if errCode == 11000 {
				return ErrEscrowExists
			}
		}
		return err
	}
	return nil
}

// CapacityReservationPaymentInfoCreate creates the reservation payment information
func CapacityReservationPaymentInfoCreate(ctx context.Context, db *mongo.Database, reservationPaymentInfo CapacityReservationPaymentInformation) error {
	col := db.Collection(CapacityEscrowCollection)
	_, err := col.InsertOne(ctx, reservationPaymentInfo)
	if err != nil {
		if merr, ok := err.(mongo.WriteException); ok {
			errCode := merr.WriteErrors[0].Code
			if errCode == 11000 {
				return ErrEscrowExists
			}
		}
		return err
	}
	return nil
}

// ReservationPaymentInfoUpdate update reservation payment info
func ReservationPaymentInfoUpdate(ctx context.Context, db *mongo.Database, update ReservationPaymentInformation) error {
	filter := bson.M{"_id": update.ReservationID}
	// actually update the user with final data
	if _, err := db.Collection(EscrowCollection).UpdateOne(ctx, filter, bson.M{"$set": update}); err != nil {
		return err
	}

	return nil
}

// CapacityReservationPaymentInfoUpdate update reservation payment info
func CapacityReservationPaymentInfoUpdate(ctx context.Context, db *mongo.Database, update CapacityReservationPaymentInformation) error {
	filter := bson.M{"_id": update.ReservationID}
	// actually update the user with final data
	if _, err := db.Collection(CapacityEscrowCollection).UpdateOne(ctx, filter, bson.M{"$set": update}); err != nil {
		return err
	}

	return nil
}

// ReservationPaymentInfoGet a single reservation escrow info using its id
func ReservationPaymentInfoGet(ctx context.Context, db *mongo.Database, id schema.ID) (ReservationPaymentInformation, error) {
	col := db.Collection(EscrowCollection)
	var rpi ReservationPaymentInformation
	res := col.FindOne(ctx, bson.M{"_id": id})
	if err := res.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return rpi, ErrEscrowNotFound
		}
		return rpi, err
	}
	err := res.Decode(&rpi)
	return rpi, err
}

// CapacityReservationPaymentInfoGet a single reservation escrow info using its id
func CapacityReservationPaymentInfoGet(ctx context.Context, db *mongo.Database, id schema.ID) (CapacityReservationPaymentInformation, error) {
	col := db.Collection(CapacityEscrowCollection)
	var rpi CapacityReservationPaymentInformation
	res := col.FindOne(ctx, bson.M{"_id": id})
	if err := res.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return rpi, ErrEscrowNotFound
		}
		return rpi, err
	}
	err := res.Decode(&rpi)
	return rpi, err
}

// GetAllActiveReservationPaymentInfos get all active reservation payment information
func GetAllActiveReservationPaymentInfos(ctx context.Context, db *mongo.Database) ([]ReservationPaymentInformation, error) {
	filter := bson.M{"paid": false, "expiration": bson.M{"$gt": schema.Date{Time: time.Now()}}}
	cursor, err := db.Collection(EscrowCollection).Find(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cursor over active payment infos")
	}
	paymentInfos := make([]ReservationPaymentInformation, 0)
	err = cursor.All(ctx, &paymentInfos)
	if err != nil {
		err = errors.Wrap(err, "failed to decode active payment information")
	}
	return paymentInfos, err
}

// GetAllActiveCapacityReservationPaymentInfos get all active reservation payment information
func GetAllActiveCapacityReservationPaymentInfos(ctx context.Context, db *mongo.Database) ([]CapacityReservationPaymentInformation, error) {
	filter := bson.M{"paid": false, "expiration": bson.M{"$gt": schema.Date{Time: time.Now()}}}
	cursor, err := db.Collection(CapacityEscrowCollection).Find(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cursor over active capacity payment infos")
	}
	paymentInfos := make([]CapacityReservationPaymentInformation, 0)
	err = cursor.All(ctx, &paymentInfos)
	if err != nil {
		err = errors.Wrap(err, "failed to decode active capacity payment information")
	}
	return paymentInfos, err
}

// GetAllExpiredReservationPaymentInfos get all expired reservation payment information
func GetAllExpiredReservationPaymentInfos(ctx context.Context, db *mongo.Database) ([]ReservationPaymentInformation, error) {
	filter := bson.M{"released": false, "canceled": false, "expiration": bson.M{"$lte": schema.Date{Time: time.Now()}}}
	cursor, err := db.Collection(EscrowCollection).Find(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cursor over expired payment infos")
	}
	paymentInfos := make([]ReservationPaymentInformation, 0)
	err = cursor.All(ctx, &paymentInfos)
	if err != nil {
		err = errors.Wrap(err, "failed to decode expired payment information")
	}
	return paymentInfos, err
}

// GetAllExpiredCapacityReservationPaymentInfos get all expired reservation payment information
func GetAllExpiredCapacityReservationPaymentInfos(ctx context.Context, db *mongo.Database) ([]CapacityReservationPaymentInformation, error) {
	filter := bson.M{"released": false, "canceled": false, "expiration": bson.M{"$lte": schema.Date{Time: time.Now()}}}
	cursor, err := db.Collection(CapacityEscrowCollection).Find(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cursor over expired payment infos")
	}
	paymentInfos := make([]CapacityReservationPaymentInformation, 0)
	err = cursor.All(ctx, &paymentInfos)
	if err != nil {
		err = errors.Wrap(err, "failed to decode expired payment information")
	}
	return paymentInfos, err
}
