package builders

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer"
	"github.com/threefoldtech/tfexplorer/client"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	wrklds "github.com/threefoldtech/tfexplorer/pkg/workloads"
	"github.com/threefoldtech/tfexplorer/schema"
	"github.com/threefoldtech/zos/pkg"
	"github.com/threefoldtech/zos/pkg/crypto"
)

// ReservationBuilder is a struct that can build reservations
type ReservationBuilder struct {
	reservation workloads.Reservation
	explorer    *client.Client
	userID      *tfexplorer.UserIdentity
	currencies  []string
	dryRun      bool
}

// NewReservationBuilder creates a new ReservationBuilder
func NewReservationBuilder(explorer *client.Client, userID *tfexplorer.UserIdentity) *ReservationBuilder {
	reservation := workloads.Reservation{}
	return &ReservationBuilder{
		reservation: reservation,
		explorer:    explorer,
		userID:      userID,
	}
}

// LoadReservationBuilder loads a reservation builder based on a file path
func LoadReservationBuilder(reader io.Reader, explorer *client.Client, userID *tfexplorer.UserIdentity) (*ReservationBuilder, error) {
	reservation := workloads.Reservation{}
	err := json.NewDecoder(reader).Decode(&reservation)
	if err != nil {
		return &ReservationBuilder{}, err
	}

	return &ReservationBuilder{
		reservation: reservation,
		explorer:    explorer,
		userID:      userID,
	}, nil
}

// Save saves the reservation builder to an IO.Writer
func (r *ReservationBuilder) Save(writer io.Writer) error {
	err := json.NewEncoder(writer).Encode(r.reservation)
	if err != nil {
		return err
	}
	return err
}

// Build returns the reservation
func (r *ReservationBuilder) Build() workloads.Reservation {
	r.reservation.Epoch = schema.Date{Time: time.Now()}
	return r.reservation
}

// Deploy deploys the reservation
func (r *ReservationBuilder) Deploy() (wrklds.ReservationCreateResponse, error) {
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
func (r *ReservationBuilder) DeleteReservation(resID int64) error {
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

// WithDryRun sets if dry run to the reservation
func (r *ReservationBuilder) WithDryRun(dryRun bool) *ReservationBuilder {
	r.dryRun = dryRun
	return r
}

// WithDuration sets the duration to the reservation
func (r *ReservationBuilder) WithDuration(duration time.Duration) *ReservationBuilder {
	timein := time.Now().Local().Add(duration)
	r.reservation.DataReservation.ExpirationReservation = schema.Date{Time: timein}
	r.reservation.DataReservation.ExpirationProvisioning = schema.Date{Time: timein}
	return r
}

// WithCurrencies sets the currencies to the reservation
func (r *ReservationBuilder) WithCurrencies(currencies []string) *ReservationBuilder {
	r.currencies = currencies
	return r
}

// AddVolume adds a volume builder to the reservation builder
func (r *ReservationBuilder) AddVolume(volume VolumeBuilder) *ReservationBuilder {
	r.reservation.DataReservation.Volumes = append(r.reservation.DataReservation.Volumes, volume.Volume)
	return r
}

// AddNetwork adds a network builder to the reservation builder
func (r *ReservationBuilder) AddNetwork(network NetworkBuilder) *ReservationBuilder {
	r.reservation.DataReservation.Networks = append(r.reservation.DataReservation.Networks, network.Network)
	return r
}

// AddZdb adds a zdb builder to the reservation builder
func (r *ReservationBuilder) AddZdb(zdb ZDBBuilder) *ReservationBuilder {
	r.reservation.DataReservation.Zdbs = append(r.reservation.DataReservation.Zdbs, zdb.ZDB)
	return r
}

// AddContainer adds a container builder to the reservation builder
func (r *ReservationBuilder) AddContainer(container ContainerBuilder) *ReservationBuilder {
	r.reservation.DataReservation.Containers = append(r.reservation.DataReservation.Containers, container.Container)
	return r
}

// AddK8s adds a k8s builder to the reservation builder
func (r *ReservationBuilder) AddK8s(k8s K8sBuilder) *ReservationBuilder {
	r.reservation.DataReservation.Kubernetes = append(r.reservation.DataReservation.Kubernetes, k8s.K8S)
	return r
}

func encryptSecret(plain, nodeID string) (string, error) {
	if len(plain) == 0 {
		return "", nil
	}

	pubkey, err := crypto.KeyFromID(pkg.StrIdentifier(nodeID))
	if err != nil {
		return "", err
	}

	encrypted, err := crypto.Encrypt([]byte(plain), pubkey)
	return hex.EncodeToString(encrypted), err
}
