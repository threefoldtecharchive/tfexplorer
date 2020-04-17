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

// Escrow are responsible for the payment flow of a reservation
type Escrow interface {
	Run(ctx context.Context) error
	RegisterReservation(reservation workloads.Reservation, supportedCurrencies []string) (types.CustomerEscrowInformation, error)
	ReservationDeployed(reservationID schema.ID)
	ReservationCanceled(reservationID schema.ID)
}

// Free implements the Escrow interface in a way that makes all reservation free
type Free struct {
	db *mongo.Database
}

// NewFree creates a new EscrowFree object
func NewFree(db *mongo.Database) *Free { return &Free{db: db} }

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

// ReservationDeployed implements the escrow interface
func (e *Free) ReservationDeployed(reservationID schema.ID) {}

// ReservationCanceled implements the escrow interface
func (e *Free) ReservationCanceled(reservationID schema.ID) {}
