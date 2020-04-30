package builders

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer"
	"github.com/threefoldtech/tfexplorer/client"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	wrklds "github.com/threefoldtech/tfexplorer/pkg/workloads"
	"github.com/threefoldtech/tfexplorer/schema"
)

// ReservationClient is a client to deploy and delete reservations
type ReservationClient struct {
	reservation workloads.Reservation
	explorer    *client.Client
	userID      *tfexplorer.UserIdentity
	dryRun      bool
	currencies  []string
}

// NewReservationClient creates a new reservation client
func NewReservationClient(explorer *client.Client, userID *tfexplorer.UserIdentity, dryRun bool, currencies []string) *ReservationClient {
	return &ReservationClient{
		explorer:   explorer,
		userID:     userID,
		dryRun:     dryRun,
		currencies: currencies,
	}
}

// Deploy deploys the reservation
func (r *ReservationClient) Deploy() (wrklds.ReservationCreateResponse, error) {
	userID := int64(r.userID.ThreebotID)
	signer, err := client.NewSigner(r.userID.Key().PrivateKey.Seed())
	if err != nil {
		return wrklds.ReservationCreateResponse{}, errors.Wrap(err, "could not load signer")
	}

	r.reservation.CustomerTid = userID
	// we always allow user to delete his own reservations
	r.reservation.DataReservation.SigningRequestDelete.QuorumMin = 1
	r.reservation.DataReservation.SigningRequestDelete.Signers = []int64{userID}

	// set allowed the currencies as provided by the user
	r.reservation.DataReservation.Currencies = r.currencies

	bytes, err := json.Marshal(r.reservation.DataReservation)
	if err != nil {
		return wrklds.ReservationCreateResponse{}, err
	}

	r.reservation.Json = string(bytes)
	_, signature, err := signer.SignHex(r.reservation.Json)
	if err != nil {
		return wrklds.ReservationCreateResponse{}, errors.Wrap(err, "failed to sign the reservation")
	}

	r.reservation.CustomerSignature = signature

	if r.dryRun {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return wrklds.ReservationCreateResponse{}, enc.Encode(r.reservation)
	}

	response, err := r.explorer.Workloads.Create(r.reservation)
	if err != nil {
		return wrklds.ReservationCreateResponse{}, errors.Wrap(err, "failed to send reservation")
	}

	return response, nil
}

// DeleteReservation deletes a reservation by id
func (r *ReservationClient) DeleteReservation(resID int64) error {
	userID := int64(r.userID.ThreebotID)

	reservation, err := r.explorer.Workloads.Get(schema.ID(resID))
	if err != nil {
		return errors.Wrap(err, "failed to get reservation info")
	}

	signer, err := client.NewSigner(r.userID.Key().PrivateKey.Seed())
	if err != nil {
		return errors.Wrapf(err, "failed to load signer")
	}

	_, signature, err := signer.SignHex(resID, reservation.Json)
	if err != nil {
		return errors.Wrap(err, "failed to sign the reservation")
	}

	if err := r.explorer.Workloads.SignDelete(schema.ID(resID), schema.ID(userID), signature); err != nil {
		return errors.Wrapf(err, "failed to sign deletion of reservation: %d", resID)
	}

	fmt.Printf("Reservation %v marked as to be deleted\n", resID)
	return nil
}
