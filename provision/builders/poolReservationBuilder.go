package builders

import (
	"encoding/json"
	"io"

	"github.com/threefoldtech/tfexplorer/pkg/capacity"
)

// CapacityReservationBuilder is a struct that can build reservations
type CapacityReservationBuilder struct {
	reservation capacity.Reservation
}

// NewCapacityReservationBuilder creates a new CapacityReservationBuilder
func NewCapacityReservationBuilder() *CapacityReservationBuilder {
	reservation := capacity.Reservation{}
	return &CapacityReservationBuilder{
		reservation: reservation,
	}
}

// LoadCapacityReservationBuilder loads a reservation builder based on a file path
func LoadCapacityReservationBuilder(reader io.Reader) (*CapacityReservationBuilder, error) {
	reservation := capacity.Reservation{}
	err := json.NewDecoder(reader).Decode(&reservation)
	if err != nil {
		return &CapacityReservationBuilder{}, err
	}

	return &CapacityReservationBuilder{
		reservation: reservation,
	}, nil
}

// Save saves the reservation builder to an IO.Writer
func (r *CapacityReservationBuilder) Save(writer io.Writer) error {
	err := r.reservation.Validate()
	if err != nil {
		return err
	}

	err = json.NewEncoder(writer).Encode(r.reservation)
	if err != nil {
		return err
	}
	return err
}

// Build returns the reservation
func (r *CapacityReservationBuilder) Build() capacity.Reservation {
	return r.reservation
}

// WithCustomerTID sets the customer signature to the reservation
func (r *CapacityReservationBuilder) WithCustomerTID(customerTid int64) *CapacityReservationBuilder {
	r.reservation.CustomerTid = customerTid
	return r
}

// WithCustomerSignature sets the customer signature to the reservation
func (r *CapacityReservationBuilder) WithCustomerSignature(customerSignature string) *CapacityReservationBuilder {
	r.reservation.CustomerSignature = customerSignature
	return r
}

// WithPoolID sets the customer signature to the reservation
func (r *CapacityReservationBuilder) WithPoolID(poolID int64) *CapacityReservationBuilder {
	r.reservation.DataReservation.PoolID = poolID
	return r
}

// WithCUs sets the cus to the reservation
func (r *CapacityReservationBuilder) WithCUs(cus uint64) *CapacityReservationBuilder {
	r.reservation.DataReservation.CUs = cus
	return r
}

// WithSUs sets the cus to the reservation
func (r *CapacityReservationBuilder) WithSUs(sus uint64) *CapacityReservationBuilder {
	r.reservation.DataReservation.SUs = sus
	return r
}

// WithNodeIDs sets the node ids to the reservation
func (r *CapacityReservationBuilder) WithNodeIDs(nodeIDs []string) *CapacityReservationBuilder {
	r.reservation.DataReservation.NodeIDs = nodeIDs
	return r
}

// WithCurrencies sets the currencies to the reservation
func (r *CapacityReservationBuilder) WithCurrencies(currencies []string) *CapacityReservationBuilder {
	r.reservation.DataReservation.Currencies = currencies
	return r
}
