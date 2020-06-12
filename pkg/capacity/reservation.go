package capacity

type (
	// PoolReservation is a reservation type to create a new capacity pool, or
	// to add capacity to an existing pool. Ownership of a new pool is tied to
	// the person who signs the request to set up the pool. Once a pool is
	// set up, anyone can top it up with additional capacity. Even though there
	// is no restriction of who can add capacity to the pool, only the owner
	// can assign workloads to it.
	//
	// This type is based on the `generated.Reservation` type. The main reason
	// for this being a separate type, is that the aforementioned type is actually
	// not flexible, and strongly tied to regular workloads. One solution would
	// be to make a pool a type of workload, but this would be a serious hack.
	//
	// Furthermore, note that some fields have been stripped. Reason is, that a
	// capacity pool is only meant to serve as an abstract concept, internal to
	// the explorer, and later the farmer threebot. As such, there are no dedicated
	// signature fields. Other workload specific info is also stripped. Note that
	// the way of signing the reservation is kept. While this method is questionable
	// to say the least, it does mean that we will have a much easier time if we
	// decide to merge the 2 reservation types in the future, which we should still
	// do.
	PoolReservation struct {
		JSON              string          `json:"json"`
		DataReservation   ReservationData `json:"data_reservation"`
		CustomerTid       int64           `json:"customer_tid"`
		CustomerSignature string          `json:"customer_signature"`
	}

	// ReservationData is the actual data sent in a capacity pool reservation. If
	// PoolID is a non-zero value, this reservation will add the requested capacity
	// to the existing pool with the given ID.
	//
	// Although CU and SU values for workloads can be (and likely will be) floating
	// points, we only allow purchasing full units. Since such a unit is actually
	// very small, this is not a problem for over purchasing, and it simplifies
	// some stuff on our end.
	ReservationData struct {
		PoolID uint64 `json:"pool_id"`
		CUs    uint64 `json:"cus"`
		SUs    uint64 `json:"sus"`
	}
)
