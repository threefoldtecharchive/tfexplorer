package builders

import (
	"encoding/hex"
	"encoding/json"
	"io"
	"time"

	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	"github.com/threefoldtech/tfexplorer/schema"
	"github.com/threefoldtech/zos/pkg"
	"github.com/threefoldtech/zos/pkg/crypto"
)

// ReservationBuilder is a struct that can build reservations
type ReservationBuilder struct {
	reservation workloads.Reservation
}

// NewReservationBuilder creates a new ReservationBuilder
func NewReservationBuilder() *ReservationBuilder {
	reservation := workloads.Reservation{}
	return &ReservationBuilder{
		reservation: reservation,
	}
}

// LoadReservationBuilder loads a reservation builder based on a file path
func LoadReservationBuilder(reader io.Reader) (*ReservationBuilder, error) {
	reservation := workloads.Reservation{}
	err := json.NewDecoder(reader).Decode(&reservation)
	if err != nil {
		return &ReservationBuilder{}, err
	}

	return &ReservationBuilder{
		reservation: reservation,
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

// WithDuration sets the duration to the reservation
func (r *ReservationBuilder) WithDuration(duration schema.Date) *ReservationBuilder {
	r.reservation.DataReservation.ExpirationReservation = duration
	return r
}

// WithExpirationProvisioning sets the expiration of the provisioning
func (r *ReservationBuilder) WithExpirationProvisioning(expiration schema.Date) *ReservationBuilder {
	r.reservation.DataReservation.ExpirationProvisioning = expiration
	return r
}

// WithSigningRequestDeleteQuorumMin sets the signing request delete quorum minimum
func (r *ReservationBuilder) WithSigningRequestDeleteQuorumMin(quorumMin int64) *ReservationBuilder {
	r.reservation.DataReservation.SigningRequestDelete.QuorumMin = quorumMin
	return r
}

// WithSigningRequestDeleteSigners sets the signing request delete signers
func (r *ReservationBuilder) WithSigningRequestDeleteSigners(signerIDs []int64) *ReservationBuilder {
	r.reservation.DataReservation.SigningRequestDelete.Signers = signerIDs
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

// AddDebug adds a debug builder to the reservation builder
func (r *ReservationBuilder) AddDebug(debug DebugBuilder) *ReservationBuilder {
	r.reservation.DataReservation.Debug = append(r.reservation.DataReservation.Debug, debug.Debug)
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
