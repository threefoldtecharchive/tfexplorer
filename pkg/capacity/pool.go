package capacity

import (
	"math"
	"time"
)

type (
	// Pool is an abstract representation of an amount of capacity purchased by
	// someone. The system must update the amount of CU and SU that is in use
	// from the pool when a workload tied to pool is created or otherwise changes
	// state. The pool is not thread safe - it is the responsibility of the consumer
	// of this struct to ensure only 1 access at a time is performed.
	Pool struct {
		// CUs and SUs are the `compute unit seconds` and `storage unit seconds`.
		// These values represent the amount left in the pool when it was last
		// updated, and do not represent current values (unless the pool just
		// got updated)
		cus float64
		sus float64

		// node ids on which this pool is applicable, only workloads deployed
		// on these nodes must be deployed in the pool.
		nodeIDs []string

		// unix timestamp when the counters where last synced. Syncing happens by
		// deducting the amount of spent CU and SU, since the last sync, from
		// the pool, and updating this field.
		lastUpdated int64

		// amount of active CU and SU tied to this pool. This is the amount of
		// CU and SU that needs to be deducted from the pool.
		activeCU float64
		activeSU float64

		// timestamp when either CU or SU expires according to the current capacity
		// still left and the capacity being used.
		emptyAt int64
		// TODO: pool id
	}
)

// AddCapacity adds new capacity to the pool
func (p *Pool) AddCapacity(CUs float64, SUs float64) {
	p.syncCurrentCapacity()
	p.cus += CUs
	p.sus += SUs
}

// AddWorkload adds the used CU and SU of a deployed workload to the currently
// active CU and SU of the pool
func (p *Pool) AddWorkload(CU float64, SU float64) {
	p.syncCurrentCapacity()
	p.activeCU += CU
	p.activeSU += SU
	p.syncPoolExpiration()
}

// RemoveWorkload remove the used CU and SU of a deployed workload to the currently
// active CU and SU of the pool
func (p *Pool) RemoveWorkload(CU float64, SU float64) {
	p.syncCurrentCapacity()
	p.activeCU -= CU
	p.activeSU -= SU
	p.syncPoolExpiration()
}

// AllowedInPool verifies that a nodeID is in the pool.
func (p *Pool) AllowedInPool(nodeID string) bool {
	for i := range p.nodeIDs {
		if p.nodeIDs[i] == nodeID {
			return true
		}
	}

	return false
}

// recalculate the current capacity in the pool
func (p *Pool) syncCurrentCapacity() {
	now := time.Now().Unix()
	timePassed := now - p.lastUpdated
	p.cus -= p.activeCU * float64(timePassed)
	p.sus -= p.activeSU * float64(timePassed)
	p.lastUpdated = now
}

// calculate when either CU or SU will be empty
func (p *Pool) syncPoolExpiration() {
	// TODO: handle case where activeCU or activeSU is 0
	// amount of seconds in the pool
	timeToCuEmpty := p.cus / p.activeCU
	timeToSuEmpty := p.sus / p.activeSU

	shortestExpiration := timeToCuEmpty
	if timeToSuEmpty < shortestExpiration {
		shortestExpiration = timeToSuEmpty
	}

	p.emptyAt = p.lastUpdated + int64(math.Floor(shortestExpiration))
}
