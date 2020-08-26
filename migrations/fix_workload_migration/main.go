// This script fix the eventual inconsistency in the data after migration of workload from reservation to workload during introduction
// of the capacity pool concepts

// signing_request_delete.signer is null => replace null with array
//   if referece is null -> empty array
//   if refrence is not null -> get old reservation and copy signing_request_delete

// signing_request_provision.signer is null => replace null with array
//   if referece is null -> empty array
//   if refrence is not null -> get old reservation and copy signing_request_provision

// signing_request_provision.signer is empty and reference is not empty => add customer_tid into the array and update quorum_min
// signing_request_delete.signer is empty and reference is not empty => add customer_tid into the array and update quorum_min

// sining_delete is null => replace null with array
// sining_provision is null => replace null with array

package main

import (
	"context"
	"flag"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/threefoldtech/tfexplorer/models/workloads"
	"github.com/threefoldtech/tfexplorer/pkg/workloads/types"
	"github.com/threefoldtech/zos/pkg/app"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func connectDB(ctx context.Context, connectionURI string) (*mongo.Client, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(connectionURI))
	if err != nil {
		return nil, err
	}

	if err := client.Connect(ctx); err != nil {
		return nil, err
	}

	return client, nil
}

func main() {
	app.Initialize()

	var (
		dbConf string
		name   string
	)

	flag.StringVar(&dbConf, "mongo", "mongodb://localhost:27017", "connection string to mongo database")
	flag.StringVar(&name, "name", "explorer", "database name")
	flag.Parse()

	ctx := context.TODO()

	client, err := connectDB(ctx, dbConf)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer client.Disconnect(ctx)
	db := client.Database(name, nil)

	fixSigningRequest(ctx, db)
	fixSigningRequestNullArray(ctx, db)
	fixSignatureWrongType(ctx, db)
}

func fixSigningRequest(ctx context.Context, db *mongo.Database) {
	wCol := db.Collection(types.WorkloadCollection)
	resCol := db.Collection(types.ReservationCollection)

	// find all workload that have been migrated
	cur, err := wCol.Find(ctx, bson.M{"reference": bson.M{"$ne": ""}})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to query database")
	}
	defer cur.Close(ctx)

	var info workloads.ReservationInfo
	for cur.Next(ctx) {
		if err := cur.Decode(&info); err != nil {
			log.Fatal().Err(err).Msg("failed to decode workload info")
		}

		sID := strings.Split(info.Reference, "-")[0]
		rID, err := strconv.ParseInt(sID, 10, 64)
		if err != nil {
			log.Fatal().Err(err).Msgf("failed to parse reference ID %v", sID)
		}
		result := resCol.FindOne(ctx, bson.M{"_id": rID})
		var reservation types.Reservation
		if err := result.Decode(&reservation); err != nil {
			log.Fatal().Err(err).Msgf("failed to decode reservation %v", rID)
		}

		// copy the singing request from their original reservation
		info.SigningRequestDelete = reservation.DataReservation.SigningRequestDelete
		info.SigningRequestProvision = reservation.DataReservation.SigningRequestProvision

		log.Info().Msgf("fix signing request of workload %v", info.ID)
		if _, err := wCol.UpdateOne(ctx, bson.M{"_id": info.ID}, bson.M{"$set": info}); err != nil {
			log.Fatal().Msgf("failed to fix signing request of workload %v", info.ID)
		}
	}
}

func fixSigningRequestNullArray(ctx context.Context, db *mongo.Database) {
	wCol := db.Collection(types.WorkloadCollection)

	cur, err := wCol.Find(ctx, bson.M{"$or": []bson.M{
		{"signing_request_delete.signers": nil},
		{"signing_request_provision.signers": nil},
	}})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to query database")
	}
	defer cur.Close(ctx)

	var info workloads.ReservationInfo
	for cur.Next(ctx) {
		if err := cur.Decode(&info); err != nil {
			log.Fatal().Err(err).Msg("failed to decode workload info")
		}

		update := false
		if info.SigningRequestDelete.Signers == nil {
			update = true
			log.Info().Msgf("fix singing_request_delete.signers null array on workload %v", info.ID)
			info.SigningRequestDelete.Signers = []int64{}
		}

		if info.SigningRequestProvision.Signers == nil {
			update = true
			log.Info().Msgf("fix singing_request_provision.signers null array on workload %v", info.ID)
			info.SigningRequestProvision.Signers = []int64{}
		}

		if update {
			if _, err := wCol.UpdateOne(ctx, bson.M{"_id": info.ID}, bson.M{"$set": info}); err != nil {
				log.Fatal().Msgf("failed to fix null array on workload %v", info.ID)
			}
		}
	}
}

func fixSignatureWrongType(ctx context.Context, db *mongo.Database) {
	wCol := db.Collection(types.WorkloadCollection)

	cur, err := wCol.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to query database")
	}
	defer cur.Close(ctx)

	var info workloads.ReservationInfo
	for cur.Next(ctx) {
		if err := cur.Decode(&info); err != nil {
			log.Fatal().Err(err).Msg("failed to decode workload info")
		}

		update := false

		if info.SignaturesDelete == nil {
			update = true
			log.Info().Msgf("fix singing_delete on workload %v", info.ID)
			info.SignaturesDelete = []workloads.SigningSignature{}
		}

		if info.SignaturesProvision == nil {
			update = true
			log.Info().Msgf("fix singing_provision on workload %v", info.ID)
			info.SignaturesProvision = []workloads.SigningSignature{}
		}

		if update {
			if _, err := wCol.UpdateOne(ctx, bson.M{"_id": info.ID}, bson.M{"$set": info}); err != nil {
				log.Fatal().Msgf("failed to fix signatures on workload %v", info.ID)
			}
		}
	}
}
