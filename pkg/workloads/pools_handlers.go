package workloads

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/mw"
	capacitytypes "github.com/threefoldtech/tfexplorer/pkg/capacity/types"
	directory "github.com/threefoldtech/tfexplorer/pkg/directory/types"
	phonebook "github.com/threefoldtech/tfexplorer/pkg/phonebook/types"
	"github.com/threefoldtech/tfexplorer/schema"
)

func (a *API) setupPool(r *http.Request) (interface{}, mw.Response) {
	defer r.Body.Close()
	var reservation capacitytypes.Reservation
	if err := json.NewDecoder(r.Body).Decode(&reservation); err != nil {
		return nil, mw.BadRequest(err)
	}

	if err := reservation.Validate(); err != nil {
		return nil, mw.BadRequest(err)
	}

	db := mw.Database(r)

	// make sure there are no duplicate node ID's
	seenNodes := make(map[string]struct{})
	for i := range reservation.DataReservation.NodeIDs {
		if _, exists := seenNodes[reservation.DataReservation.NodeIDs[i]]; exists {
			return nil, mw.Conflict(errors.New("duplicate node ID is not allowed in capacity pool"))
		}
		seenNodes[reservation.DataReservation.NodeIDs[i]] = struct{}{}
	}

	// check if all nodes belong to the same farm
	farms, err := directory.FarmsForNodes(r.Context(), db, reservation.DataReservation.NodeIDs...)
	if err != nil {
		return nil, mw.Error(err, http.StatusInternalServerError)
	}
	if len(farms) > 1 {
		return nil, mw.BadRequest(errors.New("all nodes for a capacity pool must belong to the same farm"))
	}

	isAllFree, err := isAllFreeToUse(r.Context(), reservation.DataReservation.NodeIDs, db)
	if err != nil {
		return nil, mw.Error(err, http.StatusInternalServerError)
	}

	currencies := make([]string, len(reservation.DataReservation.Currencies))
	copy(currencies, reservation.DataReservation.Currencies)

	// filter out FreeTFT if not all the nodes can be paid with freeTFT
	if !isAllFree {
		for i, c := range currencies {
			if c == freeTFT {
				currencies = append(currencies[:i], currencies[i+1:]...)
			}
		}
	}

	var filter phonebook.UserFilter
	filter = filter.WithID(schema.ID(reservation.CustomerTid))
	user, err := filter.Get(r.Context(), db)
	if err != nil {
		return nil, mw.BadRequest(errors.Wrapf(err, "cannot find user with id '%d'", reservation.CustomerTid))
	}

	if err := reservation.Verify(user.Pubkey); err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "failed to verify customer signature"))
	}

	reservation, err = capacitytypes.CapacityReservationCreate(r.Context(), db, reservation)
	if err != nil {
		return nil, mw.Error(errors.Wrap(err, "could not insert reservation in db"))
	}

	info, err := a.capacityPlanner.Reserve(reservation, currencies)
	if err != nil {
		return nil, mw.Error(err)
	}

	return CapacityPoolCreateResponse{
		ID:                reservation.ID,
		EscrowInformation: info,
	}, mw.Created()
}

func (a *API) getPool(r *http.Request) (interface{}, mw.Response) {
	idstr := mux.Vars(r)["id"]

	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		return nil, mw.BadRequest(errors.New("id must be an integer"))
	}

	pool, err := a.capacityPlanner.PoolByID(id)
	if err != nil {
		if errors.Is(err, capacitytypes.ErrPoolNotFound) {
			return nil, mw.NotFound(errors.New("capacity pool not found"))
		}
		return nil, mw.Error(err)
	}

	return pool, nil
}

func (a *API) listPools(r *http.Request) (interface{}, mw.Response) {
	ownerstr := mux.Vars(r)["owner"]

	owner, err := strconv.ParseInt(ownerstr, 10, 64)
	if err != nil {
		return nil, mw.BadRequest(errors.New("owner id must be an integer"))
	}

	pool, err := a.capacityPlanner.PoolsForOwner(owner)
	if err != nil {
		return nil, mw.Error(err)
	}

	return pool, nil
}
