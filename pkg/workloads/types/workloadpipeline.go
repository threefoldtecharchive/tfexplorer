package types

import (
	"github.com/rs/zerolog/log"
	model "github.com/threefoldtech/tfexplorer/models/workloads"
)

// WorkloadPipeline changes model.Workloader W as defined by the workload pipeline
// returns new  workload object, and true if the workload has changed
type WorkloadPipeline struct {
	w model.Workloader
}

// NewWorkloaderPipeline creates a reservation pipeline, all reservation must be processes
// through the pipeline before any action is taken. This will always make sure
// that reservation is in the right state.
func NewWorkloaderPipeline(W model.Workloader) (*WorkloadPipeline, error) {
	return &WorkloadPipeline{W}, nil
}

func (p *WorkloadPipeline) checkProvisionSignatures() bool {

	// Note: signatures validatation already done in the
	// signature add operation. Here we just make sure the
	// required quorum has been reached

	request := p.w.GetContract().SigningRequestProvision
	log.Debug().Msgf("%+v", request)
	if request.QuorumMin == 0 {
		return true
	}

	signers := countSignatures(p.w.GetState().SignaturesProvision, request)
	return int64(signers) >= request.QuorumMin
}

func (p *WorkloadPipeline) checkDeleteSignatures() bool {

	// Note: signatures validatation already done in the
	// signature add operation. Here we just make sure the
	// required quorum has been reached
	request := p.w.GetContract().SigningRequestDelete
	if request.QuorumMin == 0 {
		// if min quorum is zero, then there is no way
		// you can trigger deleting of this reservation
		return false
	}

	signers := countSignatures(p.w.GetState().SignaturesDelete, request)
	return int64(signers) >= request.QuorumMin
}

// Next gets new modified reservation, and true if the reservation has changed from the input
func (p *WorkloadPipeline) Next() (model.Workloader, bool) {
	state := p.w.GetState()
	if state.IsAny(model.NextActionDelete, model.NextActionDeleted) {
		return p.w, false
	}

	slog := log.With().Str("func", "pipeline.Next").Int64("id", int64(p.w.GetContract().ID)).Logger()

	// reseration expiration time must be checked, once expiration time is exceeded
	// the reservation must be deleted
	if p.checkDeleteSignatures() {
		// reservation has expired
		// set its status (next action) to delete
		slog.Debug().Msg("expired or to be deleted")
		state.NextAction = model.NextActionDelete
		return p.w, true
	}

	current := state.NextAction
	modified := false
	for {
		switch state.NextAction {
		case model.NextActionCreate:
			slog.Debug().Msg("ready to sign")
			state.NextAction = model.NextActionSign
		case model.NextActionSign:
			// this stage will not change unless all
			if p.checkProvisionSignatures() {
				slog.Debug().Msg("ready to pay")
				state.NextAction = model.NextActionPay
			}
		case model.NextActionPay:
			// NOTE: validation of the pools is static, and must happen when the
			// explorer receives the reservation.
			slog.Debug().Msg("reservation workloads attached to capacity pools - block until pool is confirmed to be ready")
		case model.NextActionDeploy:
			//nothing to do
			slog.Debug().Msg("let's deploy")
		}

		if current == state.NextAction {
			// no more changes in stage
			break
		}

		current = state.NextAction
		modified = true
	}

	return p.w, modified
}
