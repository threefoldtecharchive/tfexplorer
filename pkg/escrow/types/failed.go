package types

import (
	"context"

	"github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// CapacityEscrowCollection db collection for mapping between payment request and transaction sequence number and operations ids
	FailedPaymentsCollectoins = "capacity-failed-transactions"
)

type (
	FailedPaymentInfo struct {
		// ID of the pool
		ReservatoinID schema.ID `bson:"res_id"`
		// MemoText the memo text of the payment request
		MemoText string `bson:"memo_text"`
		// TxSequence the sequence number of the
		ErrorCode string `bson:"error_code"`
		// TxSequence the sequence number of the
		EnvelopeXDR string `bson:"error_code"`
		// OperationIDs list of indices in the transaction not the stellar operation id
		ResultString string `bson:"result_string"`
	}
)

// FailedPaymentInfoInfoCreate creates the new failed payment info document
func FailedPaymentInfoInfoCreate(ctx context.Context, db *mongo.Database, info FailedPaymentInfo) error {
	col := db.Collection(FailedPaymentsCollectoins)
	_, err := col.InsertOne(ctx, info)
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
