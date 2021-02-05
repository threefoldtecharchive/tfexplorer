package escrow

import (
	"context"

	capacitytypes "github.com/threefoldtech/tfexplorer/pkg/capacity/types"
	"github.com/threefoldtech/tfexplorer/pkg/escrow/types"
	"github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/mongo"
)

type (
	// Escrow are responsible for the payment flow of a reservation
	Escrow interface {
		Run(ctx context.Context) error
		CapacityReservation(reservation capacitytypes.Reservation, supportedCurrencies []string) (types.CustomerCapacityEscrowInformation, error)
		PaidCapacity() <-chan schema.ID
	}
)

// Free implements the Escrow interface in a way that makes all reservation free
type Free struct {
	db           *mongo.Database
	capacityChan chan schema.ID
}

// NewFree creates a new EscrowFree object
func NewFree(db *mongo.Database) *Free {
	return &Free{db: db, capacityChan: make(chan schema.ID)}
}

// Run implements the escrow interface
func (e *Free) Run(ctx context.Context) error {
	return nil
}

// CapacityReservation implements the escrow interface
func (e *Free) CapacityReservation(reservation capacitytypes.Reservation, _ []string) (detail types.CustomerCapacityEscrowInformation, err error) {
	// free escrow does not run in its own goroutine, so it would block and cause
	// a deadlock in the planner
	go func() {
		e.capacityChan <- reservation.ID
	}()

	return detail, nil
}

// PaidCapacity implements the escrow interface
func (e *Free) PaidCapacity() <-chan schema.ID {
	return e.capacityChan
}
