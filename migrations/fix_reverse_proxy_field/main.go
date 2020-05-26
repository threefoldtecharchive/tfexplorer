// This script is used to update the missing field from the ContainerCapacity type
// It fields the empty value with the value from the JSON field on the reservation type

package main

import (
	"context"
	"encoding/json"
	"flag"
	"reflect"

	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	"github.com/threefoldtech/tfexplorer/schema"

	"github.com/rs/zerolog/log"

	"github.com/threefoldtech/tfexplorer/pkg/workloads/types"
	"github.com/threefoldtech/zos/pkg/app"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type oldReservationData struct {
	Description             string                          `bson:"description" json:"description"`
	Currencies              []string                        `bson:"currencies" json:"currencies"`
	SigningRequestProvision workloads.SigningRequest        `bson:"signing_request_provision" json:"signing_request_provision"`
	SigningRequestDelete    workloads.SigningRequest        `bson:"signing_request_delete" json:"signing_request_delete"`
	Containers              []workloads.Container           `bson:"containers" json:"containers"`
	Volumes                 []workloads.Volume              `bson:"volumes" json:"volumes"`
	Zdbs                    []workloads.ZDB                 `bson:"zdbs" json:"zdbs"`
	Networks                []workloads.Network             `bson:"networks" json:"networks"`
	Kubernetes              []workloads.K8S                 `bson:"kubernetes" json:"kubernetes"`
	Proxies                 []workloads.GatewayProxy        `bson:"proxies" json:"proxies"`
	ReserveProxy            []workloads.GatewayReverseProxy `bson:"reverse_proxies" json:"reserve_proxies"`
	Subdomains              []workloads.GatewaySubdomain    `bson:"subdomains" json:"subdomains"`
	DomainDelegates         []workloads.GatewayDelegate     `bson:"domain_delegates" json:"domain_delegates"`
	Gateway4To6s            []workloads.Gateway4To6         `bson:"gateway4to6" json:"gateway4to6"`
	ExpirationProvisioning  schema.Date                     `bson:"expiration_provisioning" json:"expiration_provisioning"`
	ExpirationReservation   schema.Date                     `bson:"expiration_reservation" json:"expiration_reservation"`
}

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

	cur, err := col.Find(ctx, bson.D{})
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		r := types.Reservation{}
		if err := cur.Decode(&r); err != nil {
			log.Error().Err(err).Send()
			continue
		}

		var data oldReservationData

		if err := json.Unmarshal([]byte(r.Json), &data); err != nil {
			log.Fatal().Err(err).Msg("invalid json data on reservation")
		}

		if !reflect.DeepEqual(r.DataReservation, data) {
			log.Info().Msgf("start update %d", r.ID)

			r.DataReservation.ReverseProxy = data.ReserveProxy

			if !reflect.DeepEqual(r.DataReservation.ReverseProxy, data.ReserveProxy) {
				log.Error().Msgf("\n%+v\n%+v", r.DataReservation, data)
				log.Fatal().Msg("json data does not match the reservation data")
			}

			filter := bson.D{}
			filter = append(filter, bson.E{Key: "_id", Value: r.ID})

			res, err := col.UpdateOne(ctx, filter, bson.M{"$set": r})
			if err != nil {
				log.Fatal().Err(err).Msgf("failed to update %d", r.ID)
			}
			if res.ModifiedCount == 1 {
				log.Info().Msgf("updated %d", r.ID)
			} else {
				log.Error().Msgf("no document updated %d", r.ID)
			}

		}

	}
}
