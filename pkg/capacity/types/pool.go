package types

import (
	"context"
	"math"
	"math/big"
	"time"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/models"
	"github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// CapacityPoolCollection db collection name
	CapacityPoolCollection = "capacity-pools"
)

type (
	// Pool is an abstract representation of an amount of capacity purchased by
	// someone. The system must update the amount of CU and SU that is in use
	// from the pool when a workload tied to pool is created or otherwise changes
	// state. The pool is not thread safe - it is the responsibility of the consumer
	// of this struct to ensure only 1 access at a time is performed.
	Pool struct {
		// ID is the id of the pool, which needs to be referenced in workloads
		// wanting to use this pool to deploy. It can also be used to increase
		// the pool
		ID schema.ID `bson:"_id" json:"pool_id"`

		// CUs and SUs are the `compute unit seconds` and `storage unit seconds`.
		// These values represent the amount left in the pool when it was last
		// updated, and do not represent current values (unless the pool just
		// got updated)
		Cus    float64 `bson:"cus" json:"cus"`
		Sus    float64 `bson:"sus" json:"sus"`
		IPv4us float64 `bson:"ipv4us" json:"ipv4us"`
		// node ids on which this pool is applicable, only workloads deployed
		// on these nodes must be deployed in the pool.
		NodeIDs []string `bson:"node_ids" json:"node_ids"`

		// unix timestamp when the counters where last synced. Syncing happens by
		// deducting the amount of spent CU and SU, since the last sync, from
		// the pool, and updating this field.
		LastUpdated int64 `bson:"last_updated" json:"last_updated"`

		// amount of active CU and SU tied to this pool. This is the amount of
		// CU and SU that needs to be deducted from the pool.
		ActiveCU    float64 `bson:"active_cu" json:"active_cu"`
		ActiveSU    float64 `bson:"active_su" json:"active_su"`
		ActiveIPv4U float64 `bson:"active_ipv4" json:"active_ipv4"`
		// timestamp when either CU or SU expires according to the current capacity
		// still left and the capacity being used.
		EmptyAt int64 `bson:"empty_at" json:"empty_at"`

		// CustomerTid is the threebot id of the pool owner. Only the owner can
		// assign workloads to the pool
		CustomerTid int64 `bson:"customer_tid" json:"customer_tid"`
		// SponsorTid is the original sponsor of the pool when created.
		SponsorTid int64 `bson:"sponsor_tid" json:"sponsor_tid"`

		// ActiveWorkloadIDs for this pool, this list contains only unique entries
		ActiveWorkloadIDs []schema.ID `bson:"active_workload_ids" json:"active_workload_ids"`
	}
)

var (
	// ErrPoolNotFound is returned when looking for a specific pool, which is not
	// there.
	ErrPoolNotFound = errors.New("the specified pool could not be found")
	// ErrReservationNotFound is returned when a reservation with a given ID is not there
	ErrReservationNotFound = errors.New("the specified reservation was not found")
)

// NewPool sets up a new pool, ready to use, with the given data.
//
// If id is 0, it will be set on first save
func NewPool(id schema.ID, ownerID int64, sponsorTID int64, nodeIDs []string) Pool {
	return Pool{
		ID:                id,
		Cus:               0,
		Sus:               0,
		NodeIDs:           nodeIDs,
		LastUpdated:       time.Now().Unix(),
		ActiveCU:          0,
		ActiveSU:          0,
		EmptyAt:           math.MaxInt64,
		CustomerTid:       ownerID,
		SponsorTid:        sponsorTID,
		ActiveWorkloadIDs: []schema.ID{},
	}
}

// AddCapacity adds new capacity to the pool
func (p *Pool) AddCapacity(CUs float64, SUs float64, IPUs float64) {
	p.SyncCurrentCapacity()
	if CUs > 0 {
		p.Cus += CUs
	}
	if SUs > 0 {
		p.Sus += SUs
	}
	if IPUs > 0 {
		p.IPv4us += IPUs
	}

	p.syncPoolExpiration()
}

// AddWorkload adds the used CU and SU of a deployed workload to the currently
// active CU and SU of the pool, and adds the id to the actively used ids.
func (p *Pool) AddWorkload(id schema.ID, CU float64, SU float64, IPv4U float64) {
	var found bool
	for i := range p.ActiveWorkloadIDs {
		if p.ActiveWorkloadIDs[i] == id {
			found = true
			break
		}
	}
	if found {
		// workload has already been added
		return
	}
	p.SyncCurrentCapacity()
	if CU > 0 {
		p.ActiveCU += CU
	}
	if SU > 0 {
		p.ActiveSU += SU
	}
	if IPv4U > 0 {
		p.ActiveIPv4U += IPv4U
	}
	p.ActiveWorkloadIDs = append(p.ActiveWorkloadIDs, id)
	p.syncPoolExpiration()
}

// RemoveWorkload remove the used CU and SU of a deployed workload to the currently
// active CU and SU of the pool, and removes the id from the actively used id's.
func (p *Pool) RemoveWorkload(id schema.ID, CU float64, SU float64, IPv4U float64) {
	var found bool
	for i := range p.ActiveWorkloadIDs {
		if p.ActiveWorkloadIDs[i] == id {
			found = true
			// remove the id already
			// swap the element to remove with the last element, then truncate the slice
			// this messes up ordering which we don't have anyway, and is significantly
			// more performant than removing the element in the middle of the slice
			p.ActiveWorkloadIDs[len(p.ActiveWorkloadIDs)-1], p.ActiveWorkloadIDs[i] = p.ActiveWorkloadIDs[i], p.ActiveWorkloadIDs[len(p.ActiveWorkloadIDs)-1]
			p.ActiveWorkloadIDs = p.ActiveWorkloadIDs[:len(p.ActiveWorkloadIDs)-1]
			break
		}
	}
	if !found {
		// workload hasn't been added, or already removed
		return
	}
	p.SyncCurrentCapacity()
	// when clamping values also account for float skews on the positive side.
	// 0.00001 should be a small enough value that no real workload can have this
	// small amount, while it is significantly larger than what we expect from
	// float weirdness.
	if CU > 0 {
		p.ActiveCU -= CU
		// floats are annoying
		if p.ActiveCU < 0.00001 {
			p.ActiveCU = 0
		}
	}
	if SU > 0 {
		p.ActiveSU -= SU
		// floats are annoying
		if p.ActiveSU < 0.00001 {
			p.ActiveSU = 0
		}
	}
	if IPv4U > 0 {
		p.ActiveIPv4U -= IPv4U
		// floats are annoying
		if p.ActiveIPv4U < 0.00001 {
			p.ActiveIPv4U = 0
		}
	}
	p.syncPoolExpiration()
}

// AllowedInPool verifies that a nodeID is in the pool.
func (p *Pool) AllowedInPool(nodeID string) bool {
	for i := range p.NodeIDs {
		if p.NodeIDs[i] == nodeID {
			return true
		}
	}

	return false
}

// SyncCurrentCapacity recalculate the current capacity in the pool
func (p *Pool) SyncCurrentCapacity() {
	now := time.Now().Unix()
	timePassed := now - p.LastUpdated
	p.Cus -= p.ActiveCU * float64(timePassed)
	if p.Cus < 0 {
		p.Cus = 0
	}
	p.Sus -= p.ActiveSU * float64(timePassed)
	if p.Sus < 0 {
		p.Sus = 0
	}
	p.IPv4us -= p.ActiveIPv4U * float64(timePassed)
	if p.IPv4us < 0 {
		p.IPv4us = 0
	}
	p.LastUpdated = now
}

// calculate when either CU or SU will be empty
func (p *Pool) syncPoolExpiration() {
	// handle case where activeCU or activeSU is 0
	// amount of seconds in the pool
	// set expiration to the max possible length. Note that we base the initial
	// calculation off an int64 rather than a float64, since a float64 would
	// overflow the eventual int64 timestamp we calculate.
	timeToCuEmpty := big.NewFloat(0)
	// set to same precision as int64 to avoid rounding errors,
	// mainly if the active cu is 0
	timeToCuEmpty.SetPrec(64)
	if p.ActiveCU == 0 {
		timeToCuEmpty.Sub(big.NewFloat(math.MaxInt64), big.NewFloat(float64(p.LastUpdated)))
	} else {
		timeToCuEmpty.Quo(big.NewFloat(p.Cus), big.NewFloat(p.ActiveCU))
	}

	timeToSuEmpty := big.NewFloat(0)
	timeToSuEmpty.SetPrec(64)
	if p.ActiveSU == 0 {
		timeToSuEmpty.Sub(big.NewFloat(math.MaxInt64), big.NewFloat(float64(p.LastUpdated)))
	} else {
		timeToSuEmpty.Quo(big.NewFloat(p.Sus), big.NewFloat(p.ActiveSU))
	}

	timeToIpv4uEmpty := big.NewFloat(0)
	timeToIpv4uEmpty.SetPrec(64)
	if p.ActiveIPv4U == 0 {
		timeToIpv4uEmpty.Sub(big.NewFloat(math.MaxInt64), big.NewFloat(float64(p.LastUpdated)))
	} else {
		timeToIpv4uEmpty.Quo(big.NewFloat(p.IPv4us), big.NewFloat(p.ActiveIPv4U))
	}

	shortestExpiration := timeToCuEmpty
	if timeToSuEmpty.Cmp(shortestExpiration) == -1 {
		shortestExpiration = timeToSuEmpty
	}
	if timeToIpv4uEmpty.Cmp(shortestExpiration) == -1 {
		shortestExpiration = timeToIpv4uEmpty
	}

	expiration := shortestExpiration.Add(shortestExpiration, big.NewFloat(float64(p.LastUpdated)))
	xp, _ := expiration.Int64()
	p.EmptyAt = xp
}

// CapacityPoolCreate save new capacity pool to the database
func CapacityPoolCreate(ctx context.Context, db *mongo.Database, pool Pool) (Pool, error) {
	if pool.ID == 0 {
		pool.ID = models.MustID(ctx, db, CapacityReservationCollection)
	}
	_, err := db.Collection(CapacityPoolCollection).InsertOne(ctx, pool)
	if err != nil {
		return pool, err
	}

	return pool, nil
}

// UpdatePool updates the pool in the database
func UpdatePool(ctx context.Context, db *mongo.Database, pool Pool) error {
	filter := bson.M{"_id": pool.ID}

	if _, err := db.Collection(CapacityPoolCollection).UpdateOne(ctx, filter, bson.M{"$set": pool}); err != nil {
		return errors.Wrap(err, "could not update document")
	}

	return nil
}

// PoolResult wrapper object that holds errors
type PoolResult struct {
	Pool
	Err error
}

// GetPools gets all pools from the database
func GetPools(ctx context.Context, db *mongo.Database) (<-chan PoolResult, error) {
	col := db.Collection(CapacityPoolCollection)

	cur, err := col.Find(ctx, options.FindOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to load pool list")
	}

	ch := make(chan PoolResult)

	go func() {
		defer close(ch)
		defer cur.Close(ctx)

		for cur.Next(ctx) {
			var pool PoolResult
			if err := cur.Decode(&pool.Pool); err != nil {
				pool.Err = err
			}

			select {
			case ch <- pool:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, err
}

// GetPool from the database with the given ID
func GetPool(ctx context.Context, db *mongo.Database, id schema.ID) (Pool, error) {
	col := db.Collection(CapacityPoolCollection)

	var pool Pool
	res := col.FindOne(ctx, bson.M{"_id": id})
	if err := res.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return pool, ErrPoolNotFound
		}
		return pool, errors.Wrap(err, "could not load pool")
	}
	err := res.Decode(&pool)
	return pool, err
}

// GetPoolsByOwner gets all pools for an owner
func GetPoolsByOwner(ctx context.Context, db *mongo.Database, owner int64) ([]Pool, error) {
	pools := []Pool{}
	cursor, err := db.Collection(CapacityPoolCollection).Find(ctx, bson.M{"customer_tid": owner})
	if err != nil {
		return nil, errors.Wrap(err, "could not load pools for owner")
	}
	if err = cursor.All(ctx, &pools); err != nil {
		return nil, errors.Wrap(err, "could not decode pools")
	}

	return pools, nil
}

// GetExpiredPools returns a list of all expired pools
func GetExpiredPools(ctx context.Context, db *mongo.Database, ts int64) ([]Pool, error) {
	pools := []Pool{}
	cursor, err := db.Collection(CapacityPoolCollection).Find(ctx, bson.M{"empty_at": bson.M{"$lte": ts}})
	if err != nil {
		return nil, errors.Wrap(err, "could not load pools for owner")
	}
	if err = cursor.All(ctx, &pools); err != nil {
		return nil, errors.Wrap(err, "could not decode pools")
	}

	return pools, nil
}

// GetNextExpiredPool gets the first pool to expire after the given timestamp
func GetNextExpiredPool(ctx context.Context, db *mongo.Database, ts int64) (Pool, error) {
	var pool Pool
	opts := options.FindOne()
	opts.Sort = bson.M{"empty_at": 1} // ascending sort based on expired at
	res := db.Collection(CapacityPoolCollection).FindOne(ctx, bson.M{"empty_at": bson.M{"$gt": ts}}, opts)
	if res.Err() != nil {
		if errors.Is(res.Err(), mongo.ErrNoDocuments) {
			return pool, ErrPoolNotFound
		}
		return pool, errors.Wrap(res.Err(), "could not fetch next pool to expire")
	}
	if err := res.Decode(&pool); err != nil {
		return pool, errors.Wrap(err, "could not decode pool")
	}

	return pool, nil
}
