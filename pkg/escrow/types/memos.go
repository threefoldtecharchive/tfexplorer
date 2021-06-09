package types

import (
	"context"

	"github.com/pkg/errors"

	"github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// CapacityEscrowCollection db collection for mapping between payment request and transaction sequence number and operations ids
	CapacityMemoTextCollection = "capacity-memo-text"
)

type (
	CapacityMemoTextInfo struct {
		// ID of the pool
		ID schema.ID `bson:"_id"`
		// MemoText the memo text of the payment request
		MemoText string `bson:"memo_text"`
		// TxSequence the sequence number of the
		TxSequence string `bson:"tx_sequence"`
		// OperationIDs list of indices in the transaction not the stellar operation id
		OperationIDs []int `bson:"operation_ids"`
	}
)

// CapacityMemoTextInfo creates the capacity memo text info document
func CapacityMemoTextInfoCreate(ctx context.Context, db *mongo.Database, info CapacityMemoTextInfo) error {
	col := db.Collection(CapacityMemoTextCollection)
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

// GetAllActiveCapacityReservationPaymentInfos get all active reservation payment information
func CapacityMemoTextInfoGet(ctx context.Context, db *mongo.Database, memo string) ([]CapacityMemoTextInfo, error) {
	filter := bson.M{"memo_text": memo}
	cursor, err := db.Collection(CapacityMemoTextCollection).Find(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cursor over capacity memo text infos")
	}
	memoInfos := make([]CapacityMemoTextInfo, 0)
	err = cursor.All(ctx, &memoInfos)
	if err != nil {
		err = errors.Wrap(err, "failed to decode capacity memo text information")
	}
	return memoInfos, err
}
