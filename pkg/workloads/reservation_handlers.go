package workloads

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfexplorer/models"
	"github.com/threefoldtech/tfexplorer/mw"
	"github.com/threefoldtech/tfexplorer/pkg/workloads/types"
)

func (a *API) get(r *http.Request) (interface{}, mw.Response) {
	id, err := a.parseID(mux.Vars(r)["res_id"])
	if err != nil {
		return nil, mw.BadRequest(fmt.Errorf("invalid reservation id"))
	}

	var filter types.ReservationFilter
	filter = filter.WithID(id)

	db := mw.Database(r)
	reservation, err := a.pipeline(filter.Get(r.Context(), db))
	if err != nil {
		return nil, mw.NotFound(err)
	}

	return reservation, nil
}

func (a *API) list(r *http.Request) (interface{}, mw.Response) {
	var filter types.ReservationFilter
	filter, err := types.ApplyQueryFilter(r, filter)
	if err != nil {
		return nil, mw.BadRequest(err)
	}

	db := mw.Database(r)
	pager := models.PageFromRequest(r)
	cur, err := filter.Find(r.Context(), db, pager)
	if err != nil {
		return nil, mw.Error(err)
	}

	defer cur.Close(r.Context())

	total, err := filter.Count(r.Context(), db)
	if err != nil {
		return nil, mw.Error(err)
	}

	reservations := []types.Reservation{}

	for cur.Next(r.Context()) {
		var reservation types.Reservation
		if err := cur.Decode(&reservation); err != nil {
			// skip reservations we can not load
			// this is probably an old reservation
			currentID := cur.Current.Lookup("_id").Int64()
			log.Error().Err(err).Int64("id", currentID).Msg("failed to decode reservation")
			continue
		}

		reservation, err := a.pipeline(reservation, nil)
		if err != nil {
			log.Error().Err(err).Int64("id", int64(reservation.ID)).Msg("failed to process reservation")
			continue
		}

		reservations = append(reservations, reservation)
	}

	pages := fmt.Sprintf("%d", models.NrPages(total, *pager.Limit))
	return reservations, mw.Ok().WithHeader("Pages", pages)
}
