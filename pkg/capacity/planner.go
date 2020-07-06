package capacity

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	"github.com/threefoldtech/tfexplorer/pkg/capacity/types"
	"github.com/threefoldtech/tfexplorer/pkg/escrow"
	escrowtypes "github.com/threefoldtech/tfexplorer/pkg/escrow/types"
	workloadtypes "github.com/threefoldtech/tfexplorer/pkg/workloads/types"
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
		Reserve(reservation types.Reservation, currencies []string) (escrowtypes.CustomerCapacityEscrowInformation, error)
		// IsAllowed checks if the pool with the given id is owned by the given
		// customer, and can deploy on the given node.
		IsAllowed(w workloads.Workloader) (bool, error)
		// HasCapacity checks if the workload could be provisioned with its attached
		// pool as it is right now.
		HasCapacity(w workloads.Workloader, seconds uint) (bool, error)
		// UpdateUsedCapacity for a pool
		UpdateUsedCapacity(w workloads.Workloader, increase bool) error
		// PoolByID returns the pool with the given ID
		PoolByID(id int64) (types.Pool, error)
		// PoolsForOwner returns all pools for a given owner
		PoolsForOwner(owner int64) ([]types.Pool, error)
	}

	// NaivePlanner simply allows all capacity purchases, and allows all workloads
	// to use a pool, as long as they both have the same owner.
	NaivePlanner struct {
		escrow escrow.Escrow

		reserveChan            chan reserveJob
		allowedChan            chan allowedJob
		hasCapacityChan        chan hasCapacityJob
		listChan               chan listPoolJob
		updateUsedCapacityChan chan updateUsedCapacityJob

		// timer when next pool is empty
		timer *time.Timer

		db  *mongo.Database
		ctx context.Context
	}

	reserveJob struct {
		reservation  types.Reservation
		currencies   []string
		responseChan chan<- reserveResponse
	}

	reserveResponse struct {
		info escrowtypes.CustomerCapacityEscrowInformation
		err  error
	}

	allowedJob struct {
		w            workloads.Workloader
		responseChan chan<- allowedResponse
	}

	allowedResponse struct {
		status bool
		err    error
	}

	hasCapacityJob struct {
		w            workloads.Workloader
		seconds      uint
		responseChan chan<- hasCapacityResponse
	}

	hasCapacityResponse struct {
		status bool
		err    error
	}

	listPoolJob struct {
		id           int64
		owner        int64
		responseChan chan<- listPoolResponse
	}

	listPoolResponse struct {
		pools []types.Pool
		err   error
	}

	updateUsedCapacityJob struct {
		w            workloads.Workloader
		used         bool
		responseChan chan<- updateUsedCapacityResponse
	}

	updateUsedCapacityResponse struct {
		err error
	}
)

const (
	unusedPoolExpiration = time.Hour * 24 * 365 * 280
)

// NewNaivePlanner creates a new NaivePlanner, using the provided escrow and
// database connection
func NewNaivePlanner(escrow escrow.Escrow, db *mongo.Database) *NaivePlanner {
	return &NaivePlanner{
		escrow:                 escrow,
		reserveChan:            make(chan reserveJob),
		allowedChan:            make(chan allowedJob),
		listChan:               make(chan listPoolJob),
		hasCapacityChan:        make(chan hasCapacityJob),
		updateUsedCapacityChan: make(chan updateUsedCapacityJob),
		db:                     db,
	}
}

// Run implements Planner
func (p *NaivePlanner) Run(ctx context.Context) {
	p.ctx = ctx

	// first make sure we decomission workloads from expired pools
	log.Info().Msg("setting up capacity planner expiration timer")
	if err := p.handlePoolExpiration(true); err != nil {
		log.Error().Err(err).Msg("failed to expire capacity pools")
	}

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("context is done, stopping planner")
		case <-p.timer.C:
			log.Info().Msg("capacity planner timer fired, pool should be expired")
			if err := p.handlePoolExpiration(true); err != nil {
				log.Error().Err(err).Msg("failure to expire capacity pool")
			}
		case job := <-p.reserveChan:
			info, err := p.reserve(job.reservation, job.currencies)
			job.responseChan <- reserveResponse{info: info, err: err}
		case job := <-p.allowedChan:
			status, err := p.isAllowed(job.w)
			job.responseChan <- allowedResponse{status: status, err: err}
		case job := <-p.hasCapacityChan:
			status, err := p.hasCapacity(job.w, job.seconds)
			job.responseChan <- hasCapacityResponse{status: status, err: err}
		case job := <-p.listChan:
			var pools []types.Pool
			var err error
			if job.id != 0 {
				var pool types.Pool
				pool, err = p.poolByID(job.id)
				pools = []types.Pool{pool}
			} else if job.owner != 0 {
				pools, err = p.poolsForOwner(job.owner)
			}
			job.responseChan <- listPoolResponse{pools: pools, err: err}
		case job := <-p.updateUsedCapacityChan:
			err := p.updateUsedCapacity(job.w, job.used)
			job.responseChan <- updateUsedCapacityResponse{err: err}
		case id := <-p.escrow.PaidCapacity():
			if err := p.addCapacity(id); err != nil {
				log.Error().Err(err).Msg("could not add capacity to pool")
			}
		}
	}
}

// Reserve implements Planner
func (p *NaivePlanner) Reserve(reservation types.Reservation, currencies []string) (escrowtypes.CustomerCapacityEscrowInformation, error) {
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
func (p *NaivePlanner) IsAllowed(w workloads.Workloader) (bool, error) {
	ch := make(chan allowedResponse)
	defer close(ch)

	p.allowedChan <- allowedJob{
		w:            w,
		responseChan: ch,
	}

	res := <-ch

	return res.status, res.err
}

// HasCapacity implements Planner
func (p *NaivePlanner) HasCapacity(w workloads.Workloader, seconds uint) (bool, error) {
	ch := make(chan hasCapacityResponse)
	defer close(ch)

	p.hasCapacityChan <- hasCapacityJob{
		w:            w,
		seconds:      seconds,
		responseChan: ch,
	}

	res := <-ch

	return res.status, res.err
}

// PoolByID implements Planner
func (p *NaivePlanner) PoolByID(id int64) (types.Pool, error) {
	ch := make(chan listPoolResponse)
	defer close(ch)

	p.listChan <- listPoolJob{
		id:           id,
		responseChan: ch,
	}

	res := <-ch
	var pool types.Pool
	if len(res.pools) > 0 {
		pool = res.pools[0]
	}
	return pool, res.err
}

// PoolsForOwner implements Planner
func (p *NaivePlanner) PoolsForOwner(owner int64) ([]types.Pool, error) {
	ch := make(chan listPoolResponse)
	defer close(ch)

	p.listChan <- listPoolJob{
		owner:        owner,
		responseChan: ch,
	}

	res := <-ch

	return res.pools, res.err
}

// UpdateUsedCapacity implements Planner
func (p *NaivePlanner) UpdateUsedCapacity(w workloads.Workloader, used bool) error {
	ch := make(chan updateUsedCapacityResponse)
	defer close(ch)

	p.updateUsedCapacityChan <- updateUsedCapacityJob{
		w:            w,
		used:         used,
		responseChan: ch,
	}

	res := <-ch

	return res.err
}

// reserve some capacity
func (p *NaivePlanner) reserve(reservation types.Reservation, currencies []string) (escrowtypes.CustomerCapacityEscrowInformation, error) {
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
		pool = types.NewPool(reservation.ID, reservation.CustomerTid, data.NodeIDs)
		pool, err = types.CapacityPoolCreate(p.ctx, p.db, pool)
		if err != nil {
			return pi, errors.Wrap(err, "could not create new capacity pool")
		}
	}

	pi, err = p.escrow.CapacityReservation(reservation, currencies)
	if err != nil {
		return pi, errors.Wrap(err, "could not set up capacity escrow")
	}

	return pi, nil
}

// isAllowed checks if the pool with the given id is owned by the user with
// the given id, and is allowed to deploy on the given nodeID
func (p *NaivePlanner) isAllowed(w workloads.Workloader) (bool, error) {
	pool, err := types.GetPool(p.ctx, p.db, schema.ID(w.GetPoolID()))
	if err != nil {
		return false, errors.Wrap(err, "could not load pool")
	}

	return pool.CustomerTid == w.GetCustomerTid() && pool.AllowedInPool(w.GetNodeID()), nil
}

// hasCapacity checks if the pool set on the workload has enough capacity to support
// the workload for the given amount of time
func (p *NaivePlanner) hasCapacity(w workloads.Workloader, seconds uint) (bool, error) {
	pool, err := types.GetPool(p.ctx, p.db, schema.ID(w.GetPoolID()))
	if err != nil {
		return false, errors.Wrap(err, "could not load pool")
	}

	pool.AddWorkload(CloudUnitsFromResourceUnits(w.GetRSU()))

	return time.Now().Add(time.Second*time.Duration(seconds)).Unix() < pool.EmptyAt, nil
}

// poolByID returns the pool with the given ID
func (p *NaivePlanner) poolByID(id int64) (types.Pool, error) {
	pool, err := types.GetPool(p.ctx, p.db, schema.ID(id))
	if err != nil {
		return types.Pool{}, errors.Wrap(err, "could not fetch pool by id")
	}
	pool.SyncCurrentCapacity()
	return pool, nil
}

// poolsForOwner lists all pools owned by the given customer
func (p *NaivePlanner) poolsForOwner(owner int64) ([]types.Pool, error) {
	pools, err := types.GetPoolsByOwner(p.ctx, p.db, owner)
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch pools for woner")
	}

	for i := range pools {
		pools[i].SyncCurrentCapacity()
	}

	return pools, nil
}

// addCapacity to a pool, and deploy all workloads linked to the pool waiting for
// pool capacity
func (p *NaivePlanner) addCapacity(id schema.ID) error {
	reservation, err := types.CapacityReservationGet(p.ctx, p.db, id)
	if err != nil {
		return errors.Wrap(err, "could not load reservation")
	}
	poolID := reservation.ID
	if reservation.DataReservation.PoolID != 0 {
		poolID = schema.ID(reservation.DataReservation.PoolID)
	}
	pool, err := types.GetPool(p.ctx, p.db, poolID)
	if err != nil {
		return errors.Wrap(err, "could not load pool")
	}

	pool.AddCapacity(float64(reservation.DataReservation.CUs), float64(reservation.DataReservation.SUs))

	if err = types.UpdatePool(p.ctx, p.db, pool); err != nil {
		return errors.Wrap(err, "could not save pool")
	}

	// load all workloads tied to this pool in pay state
	filter := workloadtypes.WorkloadFilter{}
	filter = filter.WithPoolID(int64(poolID)).WithNextAction(workloads.NextActionPay)
	workloads, err := filter.Find(p.ctx, p.db)
	if err != nil {
		return errors.Wrap(err, "could not load workloads")
	}

	for i := range workloads {
		if err = workloadtypes.WorkloadToDeploy(p.ctx, p.db, workloads[i]); err != nil {
			return errors.Wrap(err, "failed to try and deploy workload")
		}
	}

	return p.handlePoolExpiration(false)
}

func (p *NaivePlanner) updateUsedCapacity(w workloads.Workloader, used bool) error {
	pool, err := types.GetPool(p.ctx, p.db, schema.ID(w.GetPoolID()))
	if err != nil {
		return errors.Wrap(err, "could not load pool")
	}

	if used {
		pool.AddWorkload(CloudUnitsFromResourceUnits(w.GetRSU()))
	} else {
		pool.RemoveWorkload(CloudUnitsFromResourceUnits(w.GetRSU()))
	}

	if err = types.UpdatePool(p.ctx, p.db, pool); err != nil {
		errors.Wrap(err, "could not save updated pool")
	}

	return p.handlePoolExpiration(false)
}

// handlePoolExpiration sets up the planners timer to fire as soon as the next pool
// expires. If cancelOld is given, this function will check for expired pools and try
// to cancel their attached workloads.
//
// cancelOld should be false if it is known that there can not be an expired pool
// in the system. Most notably, this will be false when called from the updateUsedCapacity
// method, since that method can only change the time at which the next pool exipres,
// not expire a currently active pool.
func (p *NaivePlanner) handlePoolExpiration(cancelOld bool) error {
	// first cancel the existing timer
	if p.timer != nil {
		// do not drain the timer channel, as that could cause a deadlock if this
		// method is called because the timer expired. If this was the case, the
		// timer channel is alreay empty. If it was not the case, the new check
		// in the goroutine runtime loop will read from the new channel we set
		// later.
		p.timer.Stop()
	}

	ts := time.Now().Unix()

	if cancelOld {
		expiredPools, err := types.GetExpiredPools(p.ctx, p.db, ts)
		if err != nil {
			return errors.Wrap(err, "could not load expired pools")
		}

		for i := range expiredPools {
			filter := workloadtypes.WorkloadFilter{}.WithPoolID(int64(expiredPools[i].ID)).WithNextAction(workloadtypes.Deploy)
			workloads, err := filter.Find(p.ctx, p.db)
			if err != nil {
				return errors.Wrap(err, "could not load workloads to expire")
			}
			for j := range workloads {
				workloads[j].SetNextAction(workloadtypes.Delete)
				if err = workloadtypes.WorkloadSetNextAction(p.ctx, p.db, workloads[j].GetID(), workloadtypes.Delete); err != nil {
					return errors.Wrap(err, "could not set workload to delete state")
				}
				if err = workloadtypes.WorkloadPush(p.ctx, p.db, workloads[j]); err != nil {
					return errors.Wrap(err, "could not push workload to delete in workload queue")
				}
			}
		}
	}

	nextPoolToExpire, err := types.GetNextExpiredPool(p.ctx, p.db, ts)
	if err != nil {
		if !errors.Is(err, types.ErrPoolNotFound) {
			return errors.Wrap(err, "could not get next pool to expire")
		}
		// ErrPoolNotFound could happen if there are no pools in the system yet.
		// Since we only care for the EmptyAt field, set that to a max value of some sort
		//
		// now we can't just use math.MaxInt64. Why? because later on we use a duration
		// from this timestamp. And someone thought hey, lets make duration an alias
		// for int64. You'd be forgiven for thinking that this is all fine. But recall
		// that go represents a duration with nanoscend precision, whereas timestamp
		// is a second. See where this is going yet? According to the official go
		// doc, the largest representable duration is about 290 years. So rather
		// than calculating exactly how far in the future we can set this, we simply
		// add about 280 years to the current time.
		log.Debug().Msg("next pool to expire not found, setting expiration at maximum")
		nextPoolToExpire.EmptyAt = time.Now().Add(unusedPoolExpiration).Unix()
	}

	// clamp max interval to prevent an overflow causing weird behavior later
	//
	// once again you may wonder, why not use `time.After(...)` here? As it turns
	// out, this also does not behave properly with large timestamps, like the ones
	// we would want to clamp.
	if nextPoolToExpire.EmptyAt > time.Now().Add(unusedPoolExpiration).Unix() {
		nextPoolToExpire.EmptyAt = time.Now().Add(unusedPoolExpiration).Unix()
	}

	log.Debug().Time("ExpireAt", time.Unix(nextPoolToExpire.EmptyAt, 0)).Msg("next pool to expire")

	p.timer = time.NewTimer(time.Until(time.Unix(nextPoolToExpire.EmptyAt, 0)))

	return nil
}
