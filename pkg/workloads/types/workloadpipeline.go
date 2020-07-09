package types

import (
	"github.com/rs/zerolog/log"
	generated "github.com/threefoldtech/tfexplorer/models/generated/workloads"
)

// WorkloadPipeline changes WorkloaderType W as defined by the workload pipeline
// returns new  workload object, and true if the workload has changed
type WorkloadPipeline struct {
	w WorkloaderType
}

// NewWorkloaderPipeline creates a reservation pipeline, all reservation must be processes
// through the pipeline before any action is taken. This will always make sure
// that reservation is in the right state.
func NewWorkloaderPipeline(W WorkloaderType) (*WorkloadPipeline, error) {
	return &WorkloadPipeline{W}, nil
}

func (p *WorkloadPipeline) checkProvisionSignatures() bool {

	// Note: signatures validatation already done in the
	// signature add operation. Here we just make sure the
	// required quorum has been reached

	request := p.w.GetSigningRequestProvision()
	log.Debug().Msgf("%+v", request)
	if request.QuorumMin == 0 {
		return true
	}

	in := func(i int64, l []int64) bool {
		for _, x := range l {
			if x == i {
				return true
			}
		}
		return false
	}

	signatures := p.w.GetSignaturesProvision()
	var count int64
	for _, signature := range signatures {
		if !in(signature.Tid, request.Signers) {
			continue
		}
		count++
	}

	return count >= request.QuorumMin
}

func (p *WorkloadPipeline) checkDeleteSignatures() bool {

	// Note: signatures validatation already done in the
	// signature add operation. Here we just make sure the
	// required quorum has been reached
	request := p.w.GetSigningRequestDelete()
	if request.QuorumMin == 0 {
		// if min quorum is zero, then there is no way
		// you can trigger deleting of this reservation
		return false
	}

	in := func(i int64, l []int64) bool {
		for _, x := range l {
			if x == i {
				return true
			}
		}
		return false
	}

	signatures := p.w.GetSignaturesDelete()
	var count int64
	for _, signature := range signatures {
		if !in(signature.Tid, request.Signers) {
			continue
		}
		count++
	}

	return count >= request.QuorumMin
}

// Next gets new modified reservation, and true if the reservation has changed from the input
func (p *WorkloadPipeline) Next() (WorkloaderType, bool) {
	if p.w.GetNextAction() == generated.NextActionDelete ||
		p.w.GetNextAction() == generated.NextActionDeleted {
		return p.w, false
	}

	slog := log.With().Str("func", "pipeline.Next").Int64("id", int64(p.w.GetID())).Logger()

	// reseration expiration time must be checked, once expiration time is exceeded
	// the reservation must be deleted
	if p.checkDeleteSignatures() {
		// reservation has expired
		// set its status (next action) to delete
		slog.Debug().Msg("expired or to be deleted")
		p.w.SetNextAction(generated.NextActionDelete)
		return p.w, true
	}

	current := p.w.GetNextAction()
	modified := false
	for {
		switch p.w.GetNextAction() {
		case generated.NextActionCreate:
			slog.Debug().Msg("ready to sign")
			p.w.SetNextAction(generated.NextActionSign)
		case generated.NextActionSign:
			// this stage will not change unless all
			if p.checkProvisionSignatures() {
				slog.Debug().Msg("ready to pay")
				p.w.SetNextAction(generated.NextActionPay)
			}
		case generated.NextActionPay:
			// NOTE: validation of the pools is static, and must happen when the
			// explorer receives the reservation.
			slog.Debug().Msg("reservation workloads attached to capacity pools - block until pool is confirmed to be ready")
		case generated.NextActionDeploy:
			//nothing to do
			slog.Debug().Msg("let's deploy")
		}

		if current == p.w.GetNextAction() {
			// no more changes in stage
			break
		}

		current = p.w.GetNextAction()
		modified = true
	}

	return p.w, modified
}
