package types

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"regexp"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfexplorer/config"
	"github.com/threefoldtech/tfexplorer/models"
	generated "github.com/threefoldtech/tfexplorer/models/generated/directory"
	"github.com/threefoldtech/tfexplorer/mw"
	"github.com/threefoldtech/tfexplorer/pkg/stellar"
	"github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	farmNamePattern = regexp.MustCompile("^[a-zA-Z0-9-_]+$")
)

const (
	// FarmCollection db collection name
	FarmCollection = "farm"
)

//Farm mongo db wrapper for generated TfgridDirectoryFarm
type Farm generated.Farm

// Validate validates farm object
func (f *Farm) Validate() error {
	if !farmNamePattern.MatchString(f.Name) {
		return fmt.Errorf("invalid farm name. name can only contain alphanumeric characters dash (-) or underscore (_)")
	}

	if f.ThreebotId == 0 {
		return fmt.Errorf("threebot_id is required")
	}

	if len(f.WalletAddresses) == 0 {
		return fmt.Errorf("invalid wallet_addresses, is required")
	}

	if config.Config.Network != "" {
		found := false
		for _, a := range f.WalletAddresses {
			validator, err := stellar.NewAddressValidator(config.Config.Network, a.Asset)
			if err != nil {
				if errors.Is(err, stellar.ErrAssetCodeNotSupported) {
					continue
				}
				return errors.Wrap(err, "address validation failed")
			}

			found = true
			if err := validator.Valid(a.Address); err != nil {
				// log the internal error, then return a nice user facing error
				log.Debug().Err(err).Msg("address validation error")
				return fmt.Errorf("invalid wallet address. Please make sure you provide a valid stellar address, with a trustline for your chosen asset %s", a.Asset)
			}
		}

		if !found {
			return errors.New("no wallet found with supported asset")
		}
	}

	return nil
}

// FarmQuery helper to parse query string
type FarmQuery struct {
	FarmName string
	OwnerID  int64
}

// Parse querystring from request
func (f *FarmQuery) Parse(r *http.Request) mw.Response {
	var err error
	f.OwnerID, err = models.QueryInt(r, "owner")
	if err != nil {
		return mw.BadRequest(errors.Wrap(err, "owner should be a integer"))
	}
	f.FarmName = r.FormValue("name")
	return nil
}

// FarmFilter type
type FarmFilter bson.D

// WithID filter farm with ID
func (f FarmFilter) WithID(id schema.ID) FarmFilter {
	return append(f, bson.E{Key: "_id", Value: id})
}

// WithName filter farm with name
func (f FarmFilter) WithName(name string) FarmFilter {
	return append(f, bson.E{Key: "name", Value: name})
}

// WithOwner filter farm by owner ID
func (f FarmFilter) WithOwner(tid int64) FarmFilter {
	return append(f, bson.E{Key: "threebot_id", Value: tid})
}

// WithIP filter farm ipaddresses by ipaddress (including reservation id)
func (f FarmFilter) WithIP(ip schema.IPCidr, reservation schema.ID) FarmFilter {
	return append(f, bson.E{Key: "ipaddresses", Value: ip})
}

// WithFarmQuery filter based on FarmQuery
func (f FarmFilter) WithFarmQuery(q FarmQuery) FarmFilter {
	if len(q.FarmName) != 0 {
		f = f.WithName(q.FarmName)
	}
	if q.OwnerID != 0 {
		f = f.WithOwner(q.OwnerID)
	}
	return f

}

// Find run the filter and return a cursor result
func (f FarmFilter) Find(ctx context.Context, db *mongo.Database, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	col := db.Collection(FarmCollection)
	if f == nil {
		f = FarmFilter{}
	}

	return col.Find(ctx, f, opts...)
}

// Count number of documents matching
func (f FarmFilter) Count(ctx context.Context, db *mongo.Database) (int64, error) {
	col := db.Collection(FarmCollection)
	if f == nil {
		f = FarmFilter{}
	}

	return col.CountDocuments(ctx, f)
}

// Get one farm that matches the filter
func (f FarmFilter) Get(ctx context.Context, db *mongo.Database) (farm Farm, err error) {
	if f == nil {
		f = FarmFilter{}
	}
	col := db.Collection(FarmCollection)
	result := col.FindOne(ctx, f, options.FindOne())

	err = result.Err()
	if err != nil {
		return
	}

	err = result.Decode(&farm)
	return
}

// Delete deletes one farm that match the filter
func (f FarmFilter) Delete(ctx context.Context, db *mongo.Database) (err error) {
	if f == nil {
		f = FarmFilter{}
	}
	col := db.Collection(FarmCollection)
	_, err = col.DeleteOne(ctx, f, options.Delete())
	return err
}

// FarmCreate creates a new farm
func FarmCreate(ctx context.Context, db *mongo.Database, farm Farm) (schema.ID, error) {
	if err := farm.Validate(); err != nil {
		return 0, err
	}

	col := db.Collection(FarmCollection)
	id, err := models.NextID(ctx, db, FarmCollection)
	if err != nil {
		return id, err
	}

	farm.ID = id
	_, err = col.InsertOne(ctx, farm)
	return id, err
}

// FarmUpdate update an existing farm
func FarmUpdate(ctx context.Context, db *mongo.Database, id schema.ID, farm Farm) error {
	farm.ID = id

	if err := farm.Validate(); err != nil {
		return err
	}

	// update is a subset of Farm that only has the updatable fields.
	// this to preven the farmer from overriding other managed fields
	// like the list of IPs
	update := struct {
		ThreebotID      int64                         `bson:"threebot_id" json:"threebot_id"`
		IyoOrganization string                        `bson:"iyo_organization" json:"iyo_organization"`
		Name            string                        `bson:"name" json:"name"`
		WalletAddresses []generated.WalletAddress     `bson:"wallet_addresses" json:"wallet_addresses"`
		Location        generated.Location            `bson:"location" json:"location"`
		Email           schema.Email                  `bson:"email" json:"email"`
		ResourcePrices  []generated.NodeResourcePrice `bson:"resource_prices" json:"resource_prices"`
		PrefixZero      schema.IPRange                `bson:"prefix_zero" json:"prefix_zero"`
		GatewayIP       net.IP                        `bson:"gateway_ip" json:"gateway_ip"`
	}{
		ThreebotID:      farm.ThreebotId,
		IyoOrganization: farm.IyoOrganization,
		Name:            farm.Name,
		WalletAddresses: farm.WalletAddresses,
		Location:        farm.Location,
		Email:           farm.Email,
		ResourcePrices:  farm.ResourcePrices,
		PrefixZero:      farm.PrefixZero,
		GatewayIP:       farm.GatewayIP,
	}

	col := db.Collection(FarmCollection)
	f := FarmFilter{}.WithID(id)
	_, err := col.UpdateOne(ctx, f, bson.M{"$set": update})
	return err
}

// FarmIPReserve reserves an IP if it's only free
func FarmIPReserve(ctx context.Context, db *mongo.Database, farm schema.ID, ip schema.IPCidr, reservation schema.ID) error {
	col := db.Collection(FarmCollection)
	// filter using 0 reservation id (not reserved)
	filter := FarmFilter{}.WithID(farm).WithIP(ip, 0)

	results, err := col.UpdateOne(ctx, filter, bson.M{
		"$set": bson.M{"ipaddresses.$.reservation_id": reservation},
	})
	if err != nil {
		return err
	}

	if results.ModifiedCount != 1 {
		return fmt.Errorf("ip is not available for reservation")
	}

	return nil
}

// FarmIPRelease releases a previously reservevd IP address
func FarmIPRelease(ctx context.Context, db *mongo.Database, farm schema.ID, ip schema.IPCidr, reservation schema.ID) error {
	col := db.Collection(FarmCollection)
	// filter using 0 reservation id (not reserved)
	filter := FarmFilter{}.WithID(farm).WithIP(ip, reservation)

	results, err := col.UpdateOne(ctx, filter, bson.M{
		"$set": bson.M{"ipaddresses.$.reservation_id": 0},
	})

	if err != nil {
		return err
	}

	if results.ModifiedCount != 1 {
		return fmt.Errorf("failed to release ip address")
	}

	return nil
}

// FarmPushIP pushes ip to a farm public ips
func FarmPushIP(ctx context.Context, db *mongo.Database, id schema.ID, ip schema.IPCidr, gw net.IP) error {

	publicIP := generated.PublicIP{
		Address: ip,
		Gateway: schema.IP{gw},
	}

	if err := publicIP.Valid(); err != nil {
		return errors.Wrap(err, "invalid public ip address configuration")
	}

	col := db.Collection(FarmCollection)

	var filter FarmFilter
	filter = filter.WithID(id)
	farm, err := filter.Get(ctx, db)
	if err != nil {
		return err
	}
	for _, configuredIP := range farm.IPAddresses {
		if configuredIP.Address.String() == ip.String() {
			return nil
		}
	}

	// add IP. we have 2 pathes
	// if farm ips list is nil (or empty)
	if len(farm.IPAddresses) == 0 {
		_, err = col.UpdateOne(ctx, filter, bson.M{
			"$set": bson.M{
				"ipaddresses": []generated.PublicIP{publicIP},
			},
		})
	} else {
		// push new value
		_, err = col.UpdateOne(ctx, filter, bson.M{
			"$push": bson.M{
				"ipaddresses": publicIP,
			},
		})
	}

	return err
}

// FarmRemoveIP removes ip from a farm public ips
func FarmRemoveIP(ctx context.Context, db *mongo.Database, id schema.ID, ip schema.IPCidr) error {
	col := db.Collection(FarmCollection)
	f := FarmFilter{}.WithID(id)

	result := col.FindOne(ctx, bson.M{
		"_id":                 id,
		"ipaddresses.address": ip,
	}, &options.FindOneOptions{
		Projection: bson.M{
			"ipaddresses.$": 1,
		},
	})

	if result.Err() == mongo.ErrNoDocuments {
		return nil
	} else if result.Err() != nil {
		return result.Err()
	}
	var farm Farm
	if err := result.Decode(&farm); err != nil {
		return errors.Wrap(err, "failed to load farm")
	}

	if len(farm.IPAddresses) != 1 {
		return fmt.Errorf("invalid number of IPs returned, expecting 1 got %d", len(farm.IPAddresses))
	}

	address := farm.IPAddresses[0]
	if address.ReservationID != 0 {
		return fmt.Errorf("ip address '%s' is reserved", address.Address.String())
	}

	// TODO: what should we return in case the IP is configured but reserved.
	// NOTE: this operation will ONLY delete the IP if it's not reserved (reservation_id = empty)
	_, err := col.UpdateOne(ctx, f, bson.M{
		"$pull": bson.M{
			"ipaddresses": address,
		},
	})

	return err
}
