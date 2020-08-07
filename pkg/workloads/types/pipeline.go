package types

import (
	"time"

	"github.com/rs/zerolog/log"
	generated "github.com/threefoldtech/tfexplorer/models/generated/workloads"
)

// Pipeline changes Reservation R as defined by the reservation pipeline
// returns new reservation object, and true if the reservation has changed
type Pipeline struct {
	r Reservation
}

// NewPipeline creates a reservation pipeline, all reservation must be processes
// through the pipeline before any action is taken. This will always make sure
// that reservation is in the right state.
func NewPipeline(R Reservation) (*Pipeline, error) {
	return &Pipeline{R}, nil
}

func (p *Pipeline) checkProvisionSignatures() bool {

	// Note: signatures validatation already done in the
	// signature add operation. Here we just make sure the
	// required quorum has been reached

	request := p.r.DataReservation.SigningRequestProvision
	log.Debug().Msgf("%+v", request)
	if request.QuorumMin == 0 {
		return true
	}

	signers := countSignatures(p.r.SignaturesProvision, request)
	return int64(signers) >= request.QuorumMin
}

func (p *Pipeline) checkDeleteSignatures() bool {

	// Note: signatures validatation already done in the
	// signature add operation. Here we just make sure the
	// required quorum has been reached
	request := p.r.DataReservation.SigningRequestDelete
	if request.QuorumMin == 0 {
		// if min quorum is zero, then there is no way
		// you can trigger deleting of this reservation
		return false
	}

	signers := countSignatures(p.r.SignaturesDelete, request)
	return int64(signers) >= request.QuorumMin
}

// Next gets new modified reservation, and true if the reservation has changed from the input
func (p *Pipeline) Next() (Reservation, bool) {
	if p.r.NextAction == generated.NextActionDelete ||
		p.r.NextAction == generated.NextActionDeleted {
		return p.r, false
	}

	slog := log.With().Str("func", "pipeline.Next").Int64("id", int64(p.r.ID)).Logger()

	// reseration expiration time must be checked, once expiration time is exceeded
	// the reservation must be deleted
	if p.r.Expired() || p.checkDeleteSignatures() {
		// reservation has expired
		// set its status (next action) to delete
		slog.Debug().Msg("expired or to be deleted")
		p.r.NextAction = generated.NextActionDelete
		return p.r, true
	}

	if p.r.DataReservation.ExpirationProvisioning.Before(time.Now()) && !p.r.IsSuccessfullyDeployed() {
		log.Debug().Msg("provision expiration reached and not fully provisionned")
		p.r.NextAction = generated.NextActionDelete
		return p.r, true
	}

	current := p.r.NextAction
	modified := false
	for {
		switch p.r.NextAction {
		case generated.NextActionCreate:
			slog.Debug().Msg("ready to sign")
			p.r.NextAction = generated.NextActionSign
		case generated.NextActionSign:
			// this stage will not change unless all
			if p.checkProvisionSignatures() {
				slog.Debug().Msg("ready to pay")
				p.r.NextAction = generated.NextActionPay
			}
		case generated.NextActionPay:
			// Pay needs to block, until the escrow moves us past this point, but
			// only in case we are dealing with a deprecated style reservation.
			// Reservations who's workloads are attached to pools can deploy immediatly.
			// NOTE: validation of the pools is static, and must happen when the
			// explorer receives the reservation.
			slog.Debug().Msg("reservation workloads attached to capacity pools - continue to deploy step")
			p.r.NextAction = generated.NextActionDeploy
		case generated.NextActionDeploy:
			//nothing to do
			slog.Debug().Msg("let's deploy")
		}

		if current == p.r.NextAction {
			// no more changes in stage
			break
		}

		current = p.r.NextAction
		modified = true
	}

	return p.r, modified
}
