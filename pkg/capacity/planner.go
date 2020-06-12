package capacity

import "github.com/threefoldtech/tfexplorer/schema"

type (
	// Planner for capacity. This is the highest level manager that decides if
	// and how capacity can be used. It is also responsible for managing payment
	// for capacity.
	Planner interface {
		Reserve(reservation Reservation) (PaymentInfo, error)
	}

	// PaymentInfo for a capacity reservation
	PaymentInfo struct {
		ID     schema.ID `json:"id"`
		Memo   string    `json:"memo"`
		Amount uint64    `json:"amount"`
	}
)
