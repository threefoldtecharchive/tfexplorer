package workloads

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/models/workloads"
)

//validateReservation that the workload reservation is valid
func validateReservation(w workloads.Workloader) error {
	if w.Contract().CustomerTid == 0 {
		return fmt.Errorf("customer_tid is required")
	}

	if len(w.State().CustomerSignature) == 0 {
		return fmt.Errorf("customer_signature is required")
	}

	if len(w.Contract().Metadata) > 1024 {
		return fmt.Errorf("metadata can not be bigger than 1024 bytes")
	}

	if w.Contract().PoolID == 0 {
		return errors.New("pool is required")
	}

	if w.Contract().Reference != "" {
		return errors.New("reference is illegal for new workloads")
	}

	return nil
}
