// This script is used to rename the reserve_proxy field from the data_reservation object of a reservation
// https://github.com/threefoldtech/tfexplorer/issues/29

package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/rs/zerolog/log"

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
	col := db.Collection(types.ReservationCollection)

	result, err := col.UpdateMany(ctx, bson.D{}, bson.D{{"$rename", bson.E{"data_reservation.reserve_proxies", "data_reservation.reverse_proxies"}}}, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to rename reserve_proxy field")
	}
	fmt.Printf("%+v\n", *result)
}
