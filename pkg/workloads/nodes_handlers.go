package workloads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	model "github.com/threefoldtech/tfexplorer/models/workloads"
	"github.com/threefoldtech/tfexplorer/mw"
	"github.com/threefoldtech/tfexplorer/pkg/workloads/types"
	"github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (a *API) workloads(r *http.Request) (interface{}, mw.Response) {
	const (
		maxPageSize = 200
	)

	var (
		nodeID = mux.Vars(r)["node_id"]
	)

	db := mw.Database(r)
	workloads, err := a.queued(r.Context(), db, nodeID, maxPageSize)
	if err != nil {
		return nil, mw.Error(err)
	}
	log.Debug().Msgf("%d queue", len(workloads))

	if len(workloads) > maxPageSize {
		return workloads, nil
	}

	from, err := a.parseID(r.FormValue("from"))
	if err != nil {
		return nil, mw.BadRequest(err)
	}

	// store last reservation ID
	lastReservationID, err := types.ReservationLastID(r.Context(), db)
	if err != nil {
		return nil, mw.Error(err)
	}

	lastWorkloadID, err := types.WorkloadsLastID(r.Context(), db)
	if err != nil {
		return nil, mw.Error(err)
	}

	lastID := lastReservationID
	if lastWorkloadID > lastID {
		lastID = lastWorkloadID
	}

	filter := types.WorkloadFilter{}.WithIDGE(from)
	filter = filter.WithNodeID(nodeID)

	cur, err := filter.FindCursor(r.Context(), db)
	if err != nil {
		return nil, mw.Error(err)
	}
	defer cur.Close(r.Context())

	for cur.Next(r.Context()) {
		var w model.Codec
		if err := cur.Decode(&w); err != nil {
			return nil, mw.Error(err)
		}

		workloader, err := a.workloadpipeline(w.Workloader, nil)
		if err != nil {
			log.Error().Err(err).Int64("id", int64(workloader.GetContract().ID)).Msg("failed to process workload")
			continue
		}

		s := workloader.GetState()
		if s.NextAction == types.Delete {
			if err := types.WorkloadSetNextAction(r.Context(), db, workloader.GetContract().ID, model.NextActionDelete); err != nil {
				return nil, mw.Error(err)
			}
		}

		if !s.IsAny(types.Deploy, types.Delete) {
			continue
		}

		workloads = append(workloads, workloader)

		if len(workloads) >= maxPageSize {
			break
		}
	}

	// if we have sufficient data return
	if len(workloads) >= maxPageSize {
		return workloads, mw.Ok().WithHeader("x-last-id", fmt.Sprint(lastID))
	}

	rfilter := types.ReservationFilter{}.WithIDGE(from)
	rfilter = rfilter.WithNodeID(nodeID)

	cur, err = rfilter.Find(r.Context(), db)
	if err != nil {
		return nil, mw.Error(err)
	}

	defer cur.Close(r.Context())

	for cur.Next(r.Context()) {
		var reservation types.Reservation
		if err := cur.Decode(&reservation); err != nil {
			return nil, mw.Error(err)
		}

		reservation, err = a.pipeline(reservation, nil)
		if err != nil {
			log.Error().Err(err).Int64("id", int64(reservation.ID)).Msg("failed to process reservation")
			continue
		}

		if reservation.NextAction == types.Delete {
			if err := a.setReservationDeleted(r.Context(), db, reservation.ID); err != nil {
				return nil, mw.Error(err)
			}
		}

		// only reservations that is in right status
		if !reservation.IsAny(types.Deploy, types.Delete) {
			continue
		}

		workloads = append(workloads, reservation.Workloads(nodeID)...)

		if len(workloads) >= maxPageSize {
			break
		}
	}

	return workloads, mw.Ok().WithHeader("x-last-id", fmt.Sprint(lastID))
}

func (a *API) workloadGet(r *http.Request) (interface{}, mw.Response) {
	gwid := mux.Vars(r)["gwid"]

	rid, err := a.parseID(strings.Split(gwid, "-")[0])
	if err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "invalid reservation id part"))
	}

	var filter types.ReservationFilter
	filter = filter.WithID(rid)

	db := mw.Database(r)
	reservation, err := a.pipeline(filter.Get(r.Context(), db))
	if err != nil {
		return a.newWorkloadGet(r)
	}
	// we use an empty node-id in listing to return all workloads in this reservation
	workloads := reservation.Workloads("")

	var workload model.Workloader
	var found bool
	for _, wl := range workloads {
		if wl.GetContract().UniqueWorkloadID() == gwid {
			workload = wl
			found = true
			break
		}
	}

	if !found {
		return nil, mw.NotFound(err)
	}

	var result struct {
		model.Workloader
		Result model.Result `json:"result"`
	}

	for _, rs := range reservation.Results {
		if rs.WorkloadId == workload.GetContract().UniqueWorkloadID() {
			result.Result = rs
			break
		}
	}

	return result, nil
}

func (a *API) newWorkloadGet(r *http.Request) (interface{}, mw.Response) {
	gwid := mux.Vars(r)["gwid"]

	rid, err := a.parseID(strings.Split(gwid, "-")[0])
	if err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "invalid reservation id part"))
	}

	var filter types.WorkloadFilter
	filter = filter.WithID(rid)

	db := mw.Database(r)
	workload, err := a.workloadpipeline(filter.Get(r.Context(), db))
	if err != nil {
		return nil, mw.NotFound(err)
	}

	if workload.GetContract().UniqueWorkloadID() != gwid {
		return nil, mw.NotFound(fmt.Errorf("workload not found"))
	}

	return workload, nil
}

func (a *API) workloadPutResult(r *http.Request) (interface{}, mw.Response) {
	defer r.Body.Close()

	nodeID := mux.Vars(r)["node_id"]
	gwid := mux.Vars(r)["gwid"]

	rid, err := a.parseID(strings.Split(gwid, "-")[0])
	if err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "invalid reservation id part"))
	}

	var result model.Result
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		return nil, mw.BadRequest(err)
	}

	result.NodeId = nodeID
	result.WorkloadId = gwid
	result.Epoch = schema.Date{Time: time.Now()}

	if err := result.Verify(nodeID); err != nil {
		return nil, mw.UnAuthorized(errors.Wrap(err, "invalid result signature"))
	}

	var filter types.ReservationFilter
	filter = filter.WithID(rid)

	db := mw.Database(r)
	reservation, err := a.pipeline(filter.Get(r.Context(), db))
	if err != nil {
		return a.newworkloadPutResult(r.Context(), db, gwid, rid, result)
	}

	workloads := reservation.Workloads(nodeID)

	var found bool
	for _, wl := range workloads {
		if wl.GetContract().UniqueWorkloadID() == gwid {
			found = true
			break
		}
	}

	if !found {
		return nil, mw.NotFound(errors.New("workload not found"))
	}

	if err := types.ResultPush(r.Context(), db, rid, result); err != nil {
		return nil, mw.Error(err)
	}

	if err := types.WorkloadPop(r.Context(), db, rid); err != nil {
		return nil, mw.Error(err)
	}

	if result.State == model.ResultStateError {
		if err := a.setReservationDeleted(r.Context(), db, rid); err != nil {
			return nil, mw.Error(err)
		}
	} else if result.State == model.ResultStateOK {
		// check if entire reservation is deployed successfully
		// fetch reservation from db again to have result appended in the model
		reservation, err = a.pipeline(filter.Get(r.Context(), db))
		if err != nil {
			return nil, mw.NotFound(err)
		}

		if reservation.IsSuccessfullyDeployed() {
			a.escrow.ReservationDeployed(rid)
		}
	}

	return nil, mw.Created()
}

func (a *API) newworkloadPutResult(ctx context.Context, db *mongo.Database, gwid string, globalID schema.ID, result model.Result) (interface{}, mw.Response) {
	var filter types.WorkloadFilter
	filter = filter.WithID(globalID)

	rid, err := a.parseID(strings.Split(gwid, "-")[0])
	if err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "invalid reservation id part"))
	}

	workload, err := a.workloadpipeline(filter.Get(ctx, db))
	if err != nil {
		return nil, mw.NotFound(err)
	}

	if workload.GetContract().UniqueWorkloadID() != gwid {
		return nil, mw.NotFound(errors.New("workload id does not exist"))
	}

	if err := types.WorkloadResultPush(ctx, db, globalID, result); err != nil {
		return nil, mw.Error(err)
	}

	if err := types.WorkloadPop(ctx, db, rid); err != nil {
		return nil, mw.Error(err)
	}

	if result.State == model.ResultStateError {
		// remove capacity from pool
		if err := a.capacityPlanner.RemoveUsedCapacity(workload); err != nil {
			log.Error().Err(err).Msg("failed to decrease used capacity in pool")
			return nil, mw.Error(err)
		}
		if err := types.WorkloadSetNextAction(ctx, db, globalID, model.NextActionDelete); err != nil {
			return nil, mw.Error(err)
		}
	} else if result.State == model.ResultStateOK {
		// add capacity to pool
		if err := a.capacityPlanner.AddUsedCapacity(workload); err != nil {
			log.Error().Err(err).Msg("failed to increase used capacity in pool")
			return nil, mw.Error(err)
		}
	}
	return nil, mw.Created()
}

func (a *API) workloadPutDeleted(r *http.Request) (interface{}, mw.Response) {
	// WARNING: #TODO
	// This method does not validate the signature of the caller
	// because there is no payload in a delete call.
	// may be a simple body that has "reservation id" and "signature"
	// can be used, we use the reservation id to avoid using the same
	// request body to delete other reservations

	// HTTP Delete should not have a body though, so may be this should be
	// changed to a PUT operation.

	nodeID := mux.Vars(r)["node_id"]
	gwid := mux.Vars(r)["gwid"]

	rid, err := a.parseID(strings.Split(gwid, "-")[0])
	if err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "invalid reservation id part"))
	}

	var filter types.ReservationFilter
	filter = filter.WithID(rid)

	db := mw.Database(r)
	reservation, err := a.pipeline(filter.Get(r.Context(), db))
	if err != nil {
		return a.newworkloadPutDeleted(r.Context(), db, rid, gwid, nodeID)
	}

	workloads := reservation.Workloads(nodeID)

	var found bool
	for _, wl := range workloads {
		if wl.GetContract().UniqueWorkloadID() == gwid {
			found = true
			break
		}
	}

	if !found {
		return nil, mw.NotFound(errors.New("workload not found"))
	}

	result := reservation.ResultOf(gwid)
	if result == nil {
		// no result for this work load
		// QUESTION: should we still mark the result as deleted?
		result = &model.Result{
			WorkloadId: gwid,
			Epoch:      schema.Date{Time: time.Now()},
		}
	}

	result.State = model.ResultStateDeleted

	if err := types.ResultPush(r.Context(), db, rid, *result); err != nil {
		return nil, mw.Error(err)
	}

	if err := types.WorkloadPop(r.Context(), db, rid); err != nil {
		return nil, mw.Error(err)
	}

	// get it from store again (make sure we are up to date)
	reservation, err = a.pipeline(filter.Get(r.Context(), db))
	if err != nil {
		return nil, mw.Error(err)
	}

	if !reservation.AllDeleted() {
		return nil, nil
	}

	if err := types.ReservationSetNextAction(r.Context(), db, reservation.ID, model.NextActionDeleted); err != nil {
		return nil, mw.Error(err)
	}

	return nil, nil
}

func (a *API) newworkloadPutDeleted(ctx context.Context, db *mongo.Database, wid schema.ID, gwid string, nodeID string) (interface{}, mw.Response) {
	rid, err := a.parseID(strings.Split(gwid, "-")[0])
	if err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "invalid reservation id part"))
	}

	var filter types.WorkloadFilter
	filter = filter.WithID(wid)

	workload, err := a.workloadpipeline(filter.Get(ctx, db))
	if err != nil {
		return nil, mw.NotFound(err)
	}

	c := workload.GetContract()
	s := workload.GetState()

	if c.UniqueWorkloadID() != gwid {
		return nil, mw.NotFound(errors.New("workload not found"))
	}

	if s.Result.WorkloadId == "" {
		// no result for this work load
		// QUESTION: should we still mark the result as deleted?
		s.Result = model.Result{
			WorkloadId: gwid,
			Epoch:      schema.Date{Time: time.Now()},
		}
	}

	s.Result.State = model.ResultStateDeleted

	// remove capacity from pool
	if err := a.capacityPlanner.RemoveUsedCapacity(workload); err != nil {
		log.Error().Err(err).Msg("failed to decrease used capacity in pool")
		return nil, mw.Error(err)
	}

	if err := types.WorkloadResultPush(ctx, db, wid, s.Result); err != nil {
		return nil, mw.Error(err)
	}

	if err := types.WorkloadPop(ctx, db, rid); err != nil {
		return nil, mw.Error(err)
	}

	if err := types.WorkloadSetNextAction(ctx, db, c.ID, model.NextActionDeleted); err != nil {
		return nil, mw.Error(err)
	}

	return nil, nil
}

func (a *API) queued(ctx context.Context, db *mongo.Database, nodeID string, limit int64) ([]model.Workloader, error) {

	workloads := make([]model.Workloader, 0)

	var queue types.QueueFilter
	queue = queue.WithNodeID(nodeID)

	cur, err := queue.Find(ctx, db, options.Find().SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var wl model.Workloader
		if err := cur.Decode(&wl); err != nil {
			return nil, err
		}
		workloads = append(workloads, wl)
	}

	return workloads, nil
}
