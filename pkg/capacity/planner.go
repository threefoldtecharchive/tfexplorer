package capacity

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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
		// Run the planner
		Run(ctx context.Context)
		// Reserve some capacity
		Reserve(reservation Reservation, currencies []string) (escrowtypes.CustomerCapacityEscrowInformation, error)
		// IsAllowed checks if the pool with the given id is owned by the given
		// customer, and can deploy on the given node.
		IsAllowed(id int64, customer int64, nodeID string) (bool, error)
	}

	// NaivePlanner simply allows all capacity purchases, and allows all workloads
	// to use a pool, as long as they both have the same owner.
	NaivePlanner struct {
		escrow escrow.Escrow

		reserveChan chan reserveJob
		allowedChan chan allowedJob

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

	reserveJob struct {
		reservation  Reservation
		currencies   []string
		responseChan chan<- reserveResponse
	}

	reserveResponse struct {
		info escrowtypes.CustomerCapacityEscrowInformation
		err  error
	}

	allowedJob struct {
		poolID      int64
		customerTid int64
		nodeID      string
		responsChan chan<- allowedResponse
	}

	allowedResponse struct {
		status bool
		err    error
	}
)

// NewNaivePlanner creates a new NaivePlanner, using the provided escrow and
// database connection
func NewNaivePlanner(escrow escrow.Escrow, db *mongo.Database) *NaivePlanner {
	return &NaivePlanner{
		escrow:      escrow,
		reserveChan: make(chan reserveJob),
		allowedChan: make(chan allowedJob),
		db:          db,
	}
}

// Run implements Planner
func (p *NaivePlanner) Run(ctx context.Context) {
	p.ctx = ctx

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("context is done, stopping planner")
		case job := <-p.reserveChan:
			info, err := p.reserve(job.reservation, job.currencies)
			job.responseChan <- reserveResponse{info: info, err: err}
		case job := <-p.allowedChan:
			status, err := p.isAllowed(job.poolID, job.customerTid, job.nodeID)
			job.responsChan <- allowedResponse{status: status, err: err}
		case info := <-p.escrow.PaidCapacity():
			if err := p.addCapacity(info); err != nil {
				log.Error().Err(err).Msg("could not add capacity to pool")
			}
		}
	}
}

// Reserve implements Planner
func (p *NaivePlanner) Reserve(reservation Reservation, currencies []string) (escrowtypes.CustomerCapacityEscrowInformation, error) {
	ch := make(chan reserveResponse)
	defer close(ch)

	p.reserveChan <- reserveJob{
		reservation:  reservation,
		currencies:   currencies,
		responseChan: ch,
	}

	res := <-ch

	return res.info, res.err
}

// IsAllowed implements Planner
func (p *NaivePlanner) IsAllowed(id int64, customer int64, nodeID string) (bool, error) {
	ch := make(chan allowedResponse)
	defer close(ch)

	p.allowedChan <- allowedJob{
		poolID:      id,
		customerTid: customer,
		nodeID:      nodeID,
	}

	res := <-ch

	return res.status, res.err
}

// reserve some capacity
func (p *NaivePlanner) reserve(reservation Reservation, currencies []string) (escrowtypes.CustomerCapacityEscrowInformation, error) {
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

// isAllowed checks if the pool with the given id is owned by the user with
// the given id, and is allowed to deploy on the given nodeID
func (p *NaivePlanner) isAllowed(id int64, customer int64, nodeID string) (bool, error) {
	pool, err := types.GetPool(p.ctx, p.db, schema.ID(id))
	if err != nil {
		return false, errors.Wrap(err, "could not load pool")
	}

	return pool.CustomerTid == customer && pool.AllowedInPool(nodeID), nil
}

func (p *NaivePlanner) addCapacity(info escrowtypes.CapacityReservationInfo) error {
	pool, err := types.GetPool(p.ctx, p.db, info.ID)
	if err != nil {
		return errors.Wrap(err, "could not load pool")
	}

	pool.AddCapacity(float64(info.CUs), float64(info.SUs))

	if err = types.UpdatePool(p.ctx, p.db, pool); err != nil {
		return errors.Wrap(err, "could not save pool")
	}

	return nil
}
