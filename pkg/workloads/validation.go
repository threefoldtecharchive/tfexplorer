package workloads

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/models/workloads"
)

//validateReservation that the workload reservation is valid
func validateReservation(w workloads.Workloader) error {
	c := w.GetContract()
	s := w.GetState()
	if c.CustomerTid == 0 {
		return fmt.Errorf("customer_tid is required")
	}

	if len(s.CustomerSignature) == 0 {
		return fmt.Errorf("customer_signature is required")
	}

	if len(c.Metadata) > 1024 {
		return fmt.Errorf("metadata can not be bigger than 1024 bytes")
	}

	if c.PoolID == 0 {
		return errors.New("pool is required")
	}

	if c.Reference != "" {
		return errors.New("reference is illegal for new workloads")
	}

	return nil
}
