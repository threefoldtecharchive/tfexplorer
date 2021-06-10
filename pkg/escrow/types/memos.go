package types

import (
	"context"

	"github.com/pkg/errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// CapacityMemoTextCollection db collection for mapping between payment request and transaction sequence number and operations ids
	CapacityMemoTextCollection = "capacity-memo-text"
)

type (
	// CapacityMemoTextInfo mapping between the memo text and the transaction sequence number and operations indices in the tx
	CapacityMemoTextInfo struct {
		// MemoText the memo text of the payment request
		MemoText string `bson:"memo_text"`
		// TxSequence the sequence number of the
		TxSequence string `bson:"tx_sequence"`
		// OperationIDs list of indices in the transaction not the stellar operation id
		OperationIDs []int `bson:"operation_ids"`
	}
)

// CapacityMemoTextInfoCreate creates the capacity memo text info document
func CapacityMemoTextInfoCreate(ctx context.Context, db *mongo.Database, info CapacityMemoTextInfo) error {
	col := db.Collection(CapacityMemoTextCollection)
	_, err := col.InsertOne(ctx, info)
	if err != nil {
		return err
	}
	return nil
}

// CapacityMemoTextInfoGet get memo text transaction mapping
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
