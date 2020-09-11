package workloads

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfexplorer/models"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	generated "github.com/threefoldtech/tfexplorer/models/generated/workloads"
	"github.com/threefoldtech/tfexplorer/mw"
	"github.com/threefoldtech/tfexplorer/pkg/capacity"
	capacitytypes "github.com/threefoldtech/tfexplorer/pkg/capacity/types"
	directory "github.com/threefoldtech/tfexplorer/pkg/directory/types"
	"github.com/threefoldtech/tfexplorer/pkg/escrow"
	escrowtypes "github.com/threefoldtech/tfexplorer/pkg/escrow/types"
	phonebook "github.com/threefoldtech/tfexplorer/pkg/phonebook/types"
	"github.com/threefoldtech/tfexplorer/pkg/workloads/types"
	"github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	// API struct
	API struct {
		escrow          escrow.Escrow
		capacityPlanner capacity.Planner
	}

	// ReservationCreateResponse wraps reservation create response
	ReservationCreateResponse struct {
		ID schema.ID `json:"reservation_id"`
	}

	// CapacityPoolCreateResponse wraps capacity pool reservation create response
	CapacityPoolCreateResponse struct {
		ID                schema.ID                                     `json:"reservation_id"`
		EscrowInformation escrowtypes.CustomerCapacityEscrowInformation `json:"escrow_information,omitempty"`
	}
)

// freeTFT currency code
const freeTFT = "FreeTFT"

// minimum amount of seconds a workload needs to be able to live with a given
// pool before we even want to attempt to deploy it
const minCapacitySeconds = 120 // 2 min

func (a *API) create(r *http.Request) (interface{}, mw.Response) {
	defer r.Body.Close()

	bodyBuf := bytes.NewBuffer(nil)
	bodyBuf.ReadFrom(r.Body)
	w, err := workloads.UnmarshalJSON(bodyBuf.Bytes())
	if err != nil {
		return nil, mw.BadRequest(err)
	}

	workload := types.WorkloaderType{Workloader: w}

	// we make sure those arrays are initialized correctly
	// this will make updating the document in place much easier
	// in later stages
	workload.SetSignaturesProvision(make([]generated.SigningSignature, 0))
	workload.SetSignaturesDelete(make([]generated.SigningSignature, 0))
	workload.SetSignatureFarmer(generated.SigningSignature{})
	workload.SetResult(generated.Result{})
	workload.SetID(schema.ID(0))

	if err := workload.Validate(); err != nil {
		return nil, mw.BadRequest(err)
	}

	workload, err = a.workloadpipeline(workload, nil)
	if err != nil {
		// if failed to create pipeline, then
		// this reservation has failed initial validation
		return nil, mw.BadRequest(err)
	}

	if workload.IsAny(types.Invalid, types.Delete) {
		return nil, mw.BadRequest(fmt.Errorf("invalid request wrong status '%s'", workload.GetNextAction().String()))
	}

	db := mw.Database(r)

	var filter phonebook.UserFilter
	filter = filter.WithID(schema.ID(workload.GetCustomerTid()))
	user, err := filter.Get(r.Context(), db)
	if err != nil {
		return nil, mw.BadRequest(errors.Wrapf(err, "cannot find user with id '%d'", workload.GetCustomerTid()))
	}

	signature, err := hex.DecodeString(workload.GetCustomerSignature())
	if err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "invalid signature format, expecting hex encoded string"))
	}

	if err := workload.Verify(user.Pubkey, signature); err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "failed to verify customer signature"))
	}

	workload.SetEpoch(schema.Date{Time: time.Now()})

	allowed, err := a.capacityPlanner.IsAllowed(workload)
	if err != nil {
		if errors.Is(err, capacitytypes.ErrPoolNotFound) {
			return nil, mw.NotFound(errors.New("pool does not exist"))
		}
		log.Error().Err(err).Msg("failed to load workload capacity pool")
		return nil, mw.Error(errors.New("could not load the required capacity pool"))
	}

	if !allowed {
		return nil, mw.Forbidden(errors.New("not allowed to deploy workload on this pool"))
	}

	id, err := types.WorkloadCreate(r.Context(), db, workload)
	if err != nil {
		log.Error().Err(err).Msg("could not create workload")
		return nil, mw.Error(err)
	}

	workload, err = types.WorkloadFilter{}.WithID(id).Get(r.Context(), db)
	if err != nil {
		log.Error().Err(err).Msg("could not fetch workload we just saved")
		return nil, mw.Error(err)
	}

	allowed, err = a.capacityPlanner.HasCapacity(workload, minCapacitySeconds)
	if err != nil {
		if errors.Is(err, capacitytypes.ErrPoolNotFound) {
			log.Error().Err(err).Int64("poolID", workload.GetPoolID()).Msg("pool disappeared")
			return nil, mw.Error(errors.New("pool does not exist"))
		}
		log.Error().Err(err).Msg("failed to load workload capacity pool")
		return nil, mw.Error(errors.New("could not load the required capacity pool"))
	}

	if !allowed {
		log.Debug().Msg("don't deploy workload as its pool is almost empty")
		if err := types.WorkloadSetNextAction(r.Context(), db, id, generated.NextActionInvalid); err != nil {
			return nil, mw.Error(fmt.Errorf("failed to marked the workload as invalid:%w", err))
		}
		return ReservationCreateResponse{ID: id}, mw.PaymentRequired(errors.New("pool needs additional capacity to support this workload"))
	}

	// immediately deploy the workload
	if err := types.WorkloadToDeploy(r.Context(), db, workload); err != nil {
		log.Error().Err(err).Msg("failed to schedule the reservation to deploy")
		return nil, mw.Error(errors.New("could not schedule reservation to deploy"))
	}

	return ReservationCreateResponse{ID: id}, mw.Created()
}

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
		if errors.Is(err, capacity.ErrTransparantCapacityExtension) {
			return nil, mw.BadRequest(err)
		}
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

func (a *API) parseID(id string) (schema.ID, error) {
	v, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "invalid id format")
	}

	return schema.ID(v), nil
}

func (a *API) pipeline(r types.Reservation, err error) (types.Reservation, error) {
	if err != nil {
		return r, err
	}
	pl, err := types.NewPipeline(r)
	if err != nil {
		return r, errors.Wrap(err, "failed to process reservation state pipeline")
	}

	r, _ = pl.Next()
	return r, nil
}

func (a *API) workloadpipeline(w types.WorkloaderType, err error) (types.WorkloaderType, error) {
	if err != nil {
		return w, err
	}
	pl, err := types.NewWorkloaderPipeline(w)
	if err != nil {
		return w, errors.Wrap(err, "failed to process reservation state pipeline")
	}

	w, _ = pl.Next()
	return w, nil
}

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

func (a *API) getWorkload(r *http.Request) (interface{}, mw.Response) {
	id, err := a.parseID(mux.Vars(r)["res_id"])
	if err != nil {
		return nil, mw.BadRequest(fmt.Errorf("invalid reservation id"))
	}

	var filter types.WorkloadFilter
	filter = filter.WithID(id)

	db := mw.Database(r)
	workload, err := a.workloadpipeline(filter.Get(r.Context(), db))
	if err != nil {
		return nil, mw.NotFound(err)
	}

	return workload, nil
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

func (a *API) listWorkload(r *http.Request) (interface{}, mw.Response) {
	var filter types.WorkloadFilter
	filter, err := types.ApplyQueryFilterWorkload(r, filter)
	if err != nil {
		return nil, mw.BadRequest(err)
	}

	db := mw.Database(r)
	pager := models.PageFromRequest(r)
	cur, err := filter.FindCursor(r.Context(), db, pager)
	if err != nil {
		return nil, mw.Error(err)
	}

	defer cur.Close(r.Context())

	total, err := filter.Count(r.Context(), db)
	if err != nil {
		return nil, mw.Error(err)
	}

	reservations := []types.WorkloaderType{}

	for cur.Next(r.Context()) {
		var workload types.WorkloaderType
		if err := cur.Decode(&workload); err != nil {
			// skip reservations we can not load
			// this is probably an old reservation
			currentID := cur.Current.Lookup("_id").Int64()
			log.Error().Err(err).Int64("id", currentID).Msg("failed to decode reservation")
			continue
		}

		workload, err := a.workloadpipeline(workload, nil)
		if err != nil {
			log.Error().Err(err).Int64("id", int64(workload.GetID())).Msg("failed to process reservation")
			continue
		}

		reservations = append(reservations, workload)
	}

	pages := fmt.Sprintf("%d", models.NrPages(total, *pager.Limit))
	return reservations, mw.Ok().WithHeader("Pages", pages)
}

func (a *API) queued(ctx context.Context, db *mongo.Database, nodeID string, limit int64) ([]types.WorkloaderType, error) {

	workloads := make([]types.WorkloaderType, 0)

	var queue types.QueueFilter
	queue = queue.WithNodeID(nodeID)

	cur, err := queue.Find(ctx, db, options.Find().SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var wl types.WorkloaderType
		if err := cur.Decode(&wl); err != nil {
			return nil, err
		}
		workloads = append(workloads, wl)
	}

	return workloads, nil
}

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
		var workloader types.WorkloaderType
		if err := cur.Decode(&workloader); err != nil {
			return nil, mw.Error(err)
		}

		workloader, err = a.workloadpipeline(workloader, nil)
		if err != nil {
			log.Error().Err(err).Int64("id", int64(workloader.GetID())).Msg("failed to process workload")
			continue
		}

		if workloader.GetNextAction() == types.Delete {
			if err := types.WorkloadSetNextAction(r.Context(), db, workloader.GetID(), generated.NextActionDelete); err != nil {
				return nil, mw.Error(err)
			}
		}

		if !workloader.IsAny(types.Deploy, types.Delete) {
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

	var workload types.WorkloaderType
	var found bool
	for _, wl := range workloads {
		if wl.UniqueWorkloadID() == gwid {
			workload = wl
			found = true
			break
		}
	}

	if !found {
		return nil, mw.NotFound(err)
	}

	var result struct {
		types.WorkloaderType
		Result types.Result `json:"result"`
	}
	result.WorkloaderType = workload
	for _, rs := range reservation.Results {
		if rs.WorkloadId == workload.UniqueWorkloadID() {
			t := types.Result(rs)
			result.Result = t
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

	if workload.UniqueWorkloadID() != gwid {
		return nil, mw.NotFound(fmt.Errorf("workload not found"))
	}

	var result struct {
		types.WorkloaderType
		Result types.Result `json:"result"`
	}
	result.WorkloaderType = workload
	result.Result = types.Result(workload.GetResult())

	return result, nil
}

func (a *API) workloadPutResult(r *http.Request) (interface{}, mw.Response) {
	defer r.Body.Close()

	nodeID := mux.Vars(r)["node_id"]
	gwid := mux.Vars(r)["gwid"]

	rid, err := a.parseID(strings.Split(gwid, "-")[0])
	if err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "invalid reservation id part"))
	}

	var result types.Result
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
		if wl.UniqueWorkloadID() == gwid {
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

	if result.State == generated.ResultStateError {
		if err := a.setReservationDeleted(r.Context(), db, rid); err != nil {
			return nil, mw.Error(err)
		}
	} else if result.State == generated.ResultStateOK {
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

func (a *API) newworkloadPutResult(ctx context.Context, db *mongo.Database, gwid string, globalID schema.ID, result types.Result) (interface{}, mw.Response) {
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

	if workload.Workload().ReservationWorkload.WorkloadId != gwid {
		return nil, mw.NotFound(errors.New("workload id does not exist"))
	}

	if err := types.WorkloadResultPush(ctx, db, globalID, result); err != nil {
		return nil, mw.Error(err)
	}

	if err := types.WorkloadPop(ctx, db, rid); err != nil {
		return nil, mw.Error(err)
	}

	if result.State == generated.ResultStateError {
		// remove capacity from pool
		if err := a.capacityPlanner.RemoveUsedCapacity(workload); err != nil {
			log.Error().Err(err).Msg("failed to decrease used capacity in pool")
			return nil, mw.Error(err)
		}
		if err := types.WorkloadSetNextAction(ctx, db, globalID, generated.NextActionDelete); err != nil {
			return nil, mw.Error(err)
		}
	} else if result.State == generated.ResultStateOK {
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
		if wl.UniqueWorkloadID() == gwid {
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
		result = &types.Result{
			WorkloadId: gwid,
			Epoch:      schema.Date{Time: time.Now()},
		}
	}

	result.State = generated.ResultStateDeleted

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

	if err := types.ReservationSetNextAction(r.Context(), db, reservation.ID, generated.NextActionDeleted); err != nil {
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

	if workload.Workload().WorkloadId != gwid {
		return nil, mw.NotFound(errors.New("workload not found"))
	}

	result := workload.ResultOf(gwid)
	if result == nil {
		// no result for this work load
		// QUESTION: should we still mark the result as deleted?
		result = &types.Result{
			WorkloadId: gwid,
			Epoch:      schema.Date{Time: time.Now()},
		}
	}

	result.State = generated.ResultStateDeleted

	// remove capacity from pool
	if err := a.capacityPlanner.RemoveUsedCapacity(workload); err != nil {
		log.Error().Err(err).Msg("failed to decrease used capacity in pool")
		return nil, mw.Error(err)
	}

	if err := types.WorkloadResultPush(ctx, db, wid, *result); err != nil {
		return nil, mw.Error(err)
	}

	if err := types.WorkloadPop(ctx, db, rid); err != nil {
		return nil, mw.Error(err)
	}

	if err := types.WorkloadSetNextAction(ctx, db, workload.GetID(), generated.NextActionDeleted); err != nil {
		return nil, mw.Error(err)
	}

	return nil, nil
}

func (a *API) signProvision(r *http.Request) (interface{}, mw.Response) {
	var signature generated.SigningSignature

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, mw.BadRequest(err)
	}
	r.Body.Close() //  must close
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := json.NewDecoder(r.Body).Decode(&signature); err != nil {
		return nil, mw.BadRequest(err)
	}

	sig, err := hex.DecodeString(signature.Signature)
	if err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "invalid signature expecting hex encoded string"))
	}

	id, err := a.parseID(mux.Vars(r)["res_id"])
	if err != nil {
		return nil, mw.BadRequest(fmt.Errorf("invalid reservation id"))
	}

	var filter types.ReservationFilter
	filter = filter.WithID(id)

	db := mw.Database(r)
	reservation, err := a.pipeline(filter.Get(r.Context(), db))
	if err != nil {
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		return a.newSignProvision(r)
	}

	if reservation.NextAction != generated.NextActionSign {
		return nil, mw.UnAuthorized(fmt.Errorf("reservation not expecting signatures"))
	}

	if httpErr := userCanSign(signature.Tid, reservation.DataReservation.SigningRequestProvision, reservation.SignaturesProvision); httpErr != nil {
		return nil, httpErr
	}

	user, err := phonebook.UserFilter{}.WithID(schema.ID(signature.Tid)).Get(r.Context(), db)
	if err != nil {
		return nil, mw.NotFound(errors.Wrap(err, "customer id not found"))
	}

	if err := reservation.SignatureVerify(user.Pubkey, sig); err != nil {
		return nil, mw.UnAuthorized(errors.Wrap(err, "failed to verify signature"))
	}

	signature.Epoch = schema.Date{Time: time.Now()}
	if err := types.ReservationPushSignature(r.Context(), db, id, types.SignatureProvision, signature); err != nil {
		return nil, mw.Error(err)
	}

	reservation, err = a.pipeline(filter.Get(r.Context(), db))
	if err != nil {
		return nil, mw.Error(err)
	}

	if reservation.NextAction == generated.NextActionDeploy {
		types.WorkloadPush(r.Context(), db, reservation.Workloads("")...)
	}

	return nil, mw.Created()
}

func (a *API) newSignProvision(r *http.Request) (interface{}, mw.Response) {
	var signature generated.SigningSignature

	if err := json.NewDecoder(r.Body).Decode(&signature); err != nil {
		return nil, mw.BadRequest(err)
	}

	id, err := a.parseID(mux.Vars(r)["res_id"])
	if err != nil {
		return nil, mw.BadRequest(fmt.Errorf("invalid reservation id"))
	}

	var filter types.WorkloadFilter
	filter = filter.WithID(id)

	db := mw.Database(r)
	workload, err := a.workloadpipeline(filter.Get(r.Context(), db))
	if err != nil {
		return nil, mw.NotFound(err)
	}

	if workload.GetNextAction() != generated.NextActionSign {
		return nil, mw.UnAuthorized(fmt.Errorf("workload not expecting signatures"))
	}

	if httpErr := userCanSign(signature.Tid, workload.GetSigningRequestProvision(), workload.GetSignaturesProvision()); httpErr != nil {
		return nil, httpErr
	}

	user, err := phonebook.UserFilter{}.WithID(schema.ID(signature.Tid)).Get(r.Context(), db)
	if err != nil {
		return nil, mw.NotFound(errors.Wrap(err, "customer id not found"))
	}

	if err := workload.SignatureProvisionRequestVerify(user.Pubkey, signature); err != nil {
		return nil, mw.UnAuthorized(errors.Wrap(err, "failed to verify signature"))
	}

	signature.Epoch = schema.Date{Time: time.Now()}
	if err := types.WorkloadPushSignature(r.Context(), db, id, types.SignatureProvision, signature); err != nil {
		return nil, mw.Error(err)
	}

	workload, err = a.workloadpipeline(filter.Get(r.Context(), db))
	if err != nil {
		return nil, mw.Error(err)
	}

	if workload.GetNextAction() == generated.NextActionDeploy {
		if err = types.WorkloadPush(r.Context(), db, workload); err != nil {
			return nil, mw.Error(err)
		}
	}

	return nil, mw.Created()
}

func (a *API) signDelete(r *http.Request) (interface{}, mw.Response) {
	var signature generated.SigningSignature

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, mw.BadRequest(err)
	}
	r.Body.Close() //  must close
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := json.NewDecoder(r.Body).Decode(&signature); err != nil {
		return nil, mw.BadRequest(err)
	}

	sig, err := hex.DecodeString(signature.Signature)
	if err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "invalid signature expecting hex encoded string"))
	}

	id, err := a.parseID(mux.Vars(r)["res_id"])
	if err != nil {
		return nil, mw.BadRequest(err)
	}

	var filter types.ReservationFilter
	filter = filter.WithID(id)

	db := mw.Database(r)
	reservation, err := a.pipeline(filter.Get(r.Context(), db))
	if err != nil {
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		return a.newSignDelete(r)
	}

	if httpErr := userCanSign(signature.Tid, reservation.DataReservation.SigningRequestDelete, reservation.SignaturesDelete); httpErr != nil {
		return nil, httpErr
	}

	user, err := phonebook.UserFilter{}.WithID(schema.ID(signature.Tid)).Get(r.Context(), db)
	if err != nil {
		return nil, mw.NotFound(errors.Wrap(err, "customer id not found"))
	}

	if err := reservation.SignatureVerify(user.Pubkey, sig); err != nil {
		return nil, mw.UnAuthorized(errors.Wrap(err, "failed to verify signature"))
	}

	signature.Epoch = schema.Date{Time: time.Now()}
	if err := types.ReservationPushSignature(r.Context(), db, id, types.SignatureDelete, signature); err != nil {
		return nil, mw.Error(err)
	}

	reservation, err = a.pipeline(filter.Get(r.Context(), db))
	if err != nil {
		return nil, mw.Error(err)
	}

	if reservation.NextAction != generated.NextActionDelete {
		return nil, mw.Created()
	}

	if err := a.setReservationDeleted(r.Context(), db, reservation.ID); err != nil {
		return nil, mw.Error(err)
	}

	if err := types.WorkloadPush(r.Context(), db, reservation.Workloads("")...); err != nil {
		return nil, mw.Error(err)
	}

	return nil, mw.Created()
}

func (a *API) newSignDelete(r *http.Request) (interface{}, mw.Response) {
	var signature generated.SigningSignature

	if err := json.NewDecoder(r.Body).Decode(&signature); err != nil {
		return nil, mw.BadRequest(err)
	}

	id, err := a.parseID(mux.Vars(r)["res_id"])
	if err != nil {
		return nil, mw.BadRequest(fmt.Errorf("invalid reservation id"))
	}

	var filter types.WorkloadFilter
	filter = filter.WithID(id)

	db := mw.Database(r)
	workload, err := a.workloadpipeline(filter.Get(r.Context(), db))
	if err != nil {
		return nil, mw.NotFound(err)
	}

	if httpErr := userCanSign(signature.Tid, workload.GetSigningRequestDelete(), workload.GetSignaturesDelete()); httpErr != nil {
		return nil, httpErr
	}

	user, err := phonebook.UserFilter{}.WithID(schema.ID(signature.Tid)).Get(r.Context(), db)
	if err != nil {
		return nil, mw.NotFound(errors.Wrap(err, "customer id not found"))
	}

	if err := workload.SignatureDeleteRequestVerify(user.Pubkey, signature); err != nil {
		return nil, mw.UnAuthorized(errors.Wrap(err, "failed to verify signature"))
	}

	signature.Epoch = schema.Date{Time: time.Now()}
	if err := types.WorkloadPushSignature(r.Context(), db, id, types.SignatureDelete, signature); err != nil {
		return nil, mw.Error(err)
	}

	workload, err = a.workloadpipeline(filter.Get(r.Context(), db))
	if err != nil {
		return nil, mw.Error(err)
	}

	if workload.GetNextAction() != generated.NextActionDelete {
		return nil, mw.Created()
	}

	workload, err = a.setWorkloadDelete(r.Context(), db, workload)
	if err != nil {
		return nil, mw.Error(err)
	}

	return nil, mw.Created()
}

func (a *API) setReservationDeleted(ctx context.Context, db *mongo.Database, id schema.ID) error {
	// cancel reservation escrow in case the reservation has not yet been deployed
	a.escrow.ReservationCanceled(id)
	// No longer set the reservation as deleted. This means a workload which managed
	// to deploy will stay allive. This code path should not happen (it can only
	// happen just after the upgrade, for reservations with a pending escrow), and
	// its not worth the hassle to manually figure out where to send the tokens.
	return nil
}

func (a *API) setWorkloadDelete(ctx context.Context, db *mongo.Database, w types.WorkloaderType) (types.WorkloaderType, error) {
	w.SetNextAction(types.Delete)

	if err := types.ReservationSetNextAction(ctx, db, w.GetID(), types.Delete); err != nil {
		return w, errors.Wrap(err, "could not update workload to delete state")
	}

	return w, errors.Wrap(types.WorkloadPush(ctx, db, w), "could not push workload to delete in queue")
}

// userCanSign checks if a specific user has right to push a deletion or provision signature to the reservation/workload
func userCanSign(userTid int64, req workloads.SigningRequest, signatures []workloads.SigningSignature) mw.Response {
	in := func(i int64, l []int64) bool {
		for _, x := range l {
			if x == i {
				return true
			}
		}
		return false
	}

	// ensure the user trying to sign is required consensus
	if !in(userTid, req.Signers) {
		return mw.UnAuthorized(fmt.Errorf("signature not required for user '%d'", userTid))
	}

	// ensure the user trying to sign has not already provided a signature
	userSigned := make([]int64, 0, len(signatures))
	for i := range signatures {
		userSigned = append(userSigned, signatures[i].Tid)
	}
	if in(userTid, userSigned) {
		return mw.BadRequest(fmt.Errorf("user %d has already signed the reservation for deletion", userTid))
	}

	return nil
}

func isAllFreeToUse(ctx context.Context, nodeIDs []string, db *mongo.Database) (bool, error) {
	var freeNodes int64
	// check if freeTFT is allowed to be used
	// if all nodes are marked as free to use then FreeTFT is allowed
	// otherwise it is not
	count, err := (directory.NodeFilter{}).
		WithNodeIDs(nodeIDs).
		WithFreeToUse(true).
		ExcludeDeleted().
		Count(ctx, db)
	if err != nil {
		return false, err
	}
	freeNodes += count

	// also include the gateways belonging to the farm
	count, err = (directory.GatewayFilter{}).
		WithGWIDs(nodeIDs).
		WithFreeToUse(true).
		Count(ctx, db)
	if err != nil {
		return false, err
	}
	freeNodes += count

	log.Info().
		Int("requested_nodes", len(nodeIDs)).
		Int64("free_nodes", freeNodes).
		Msg("distribution of free nodes in capacity reservation")

	return freeNodes >= int64(len(nodeIDs)), nil
}
