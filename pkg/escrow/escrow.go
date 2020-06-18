package escrow

import (
	"context"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	"github.com/threefoldtech/tfexplorer/pkg/escrow/types"
	workloadstypes "github.com/threefoldtech/tfexplorer/pkg/workloads/types"
	"github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/mongo"
)

type (
	// Escrow are responsible for the payment flow of a reservation
	Escrow interface {
		Run(ctx context.Context) error
		RegisterReservation(reservation workloads.Reservation, supportedCurrencies []string) (types.CustomerEscrowInformation, error)
		ReservationDeployed(reservationID schema.ID)
		ReservationCanceled(reservationID schema.ID)
		CapacityReservation(reservation types.CapacityReservationInfo, customerTid int64, nodeIDs []string, supportedCurrencies []string) (types.CustomerCapacityEscrowInformation, error)
		PaidCapacity() <-chan types.CapacityReservationInfo
	}
)

// Free implements the Escrow interface in a way that makes all reservation free
type Free struct {
	db           *mongo.Database
	capacityChan chan types.CapacityReservationInfo
}

// NewFree creates a new EscrowFree object
func NewFree(db *mongo.Database) *Free {
	return &Free{db: db, capacityChan: make(chan types.CapacityReservationInfo)}
}

// Run implements the escrow interface
func (e *Free) Run(ctx context.Context) error {
	return nil
}

// RegisterReservation implements the escrow interface
func (e *Free) RegisterReservation(reservation workloads.Reservation, _ []string) (detail types.CustomerEscrowInformation, err error) {

	if reservation.NextAction == workloads.NextActionPay {
		if err = workloadstypes.ReservationSetNextAction(context.Background(), e.db, reservation.ID, workloads.NextActionDeploy); err != nil {
			err = errors.Wrapf(err, "failed to change state of reservation %d to DEPLOY", reservation.ID)
			return
		}
	}

	return detail, nil
}

// CapacityReservation implements the escrow interface
func (e *Free) CapacityReservation(info types.CapacityReservationInfo, _ int64, _ []string, _ []string) (detail types.CustomerCapacityEscrowInformation, err error) {
	// free escrow does not run in its own goroutine, so it would block and cause
	// a deadlock in the planner
	go func() {
		e.capacityChan <- info
	}()

	return detail, nil
}

// ReservationDeployed implements the escrow interface
func (e *Free) ReservationDeployed(reservationID schema.ID) {}

// ReservationCanceled implements the escrow interface
func (e *Free) ReservationCanceled(reservationID schema.ID) {}

// PaidCapacity implements the escrow interface
func (e *Free) PaidCapacity() <-chan types.CapacityReservationInfo {
	return e.capacityChan
}
