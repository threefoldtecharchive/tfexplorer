package capacity

import (
	"context"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/pkg/capacity/types"
	"github.com/threefoldtech/tfexplorer/pkg/escrow"
	escrowtypes "github.com/threefoldtech/tfexplorer/pkg/escrow/types"
	"github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/mongo"
)

type (
	// Planner for capacity. This is the highest level manager that decides if
	// and how capacity can be used. It is also responsible for managing payment
	// for capacity.
	Planner interface {
		Reserve(reservation Reservation, currencies []string) (escrowtypes.CustomerCapacityEscrowInformation, error)
	}

	// NaivePlanner simply allows all capacity purchases, and allows all workloads
	// to use a pool, as long as they both have the same owner.
	NaivePlanner struct {
		escrow escrow.Escrow

		db  *mongo.Database
		ctx context.Context
	}

	// PaymentInfo for a capacity reservation
	// TODO: DEPRECATED
	PaymentInfo struct {
		ID     schema.ID `json:"id"`
		Memo   string    `json:"memo"`
		Amount uint64    `json:"amount"`
	}
)

// Reserve implements Planner
func (p *NaivePlanner) Reserve(reservation Reservation, currencies []string) (escrowtypes.CustomerCapacityEscrowInformation, error) {
	var pi escrowtypes.CustomerCapacityEscrowInformation

	data := reservation.DataReservation

	var pool types.Pool
	var err error
	// check if we are adding to an existing pool
	if data.PoolID != 0 {
		// verify pool id
		pool, err = types.GetPool(p.ctx, p.db, schema.ID(data.PoolID))
		if err != nil {
			return pi, errors.Wrap(err, "failed to load pool")
		}

	} else {
		// create new pool
		pool = types.NewPool(reservation.CustomerTid, data.NodeIDs)
		pool, err = types.CapacityPoolCreate(p.ctx, p.db, pool)
		if err != nil {
			return pi, errors.Wrap(err, "could not create new capacity pool")
		}
	}

	// TODO: create escrow
	capInfo := escrowtypes.CapacityReservationInfo{
		ID:  pool.ID,
		CUs: data.CUs,
		SUs: data.SUs,
	}
	pi, err = p.escrow.CapacityReservation(capInfo, reservation.CustomerTid, pool.NodeIDs, currencies)
	if err != nil {
		return pi, errors.Wrap(err, "could not set up capacity escrow")
	}

	return pi, nil
}
