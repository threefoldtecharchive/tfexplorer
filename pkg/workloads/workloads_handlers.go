package workloads

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfexplorer/models"
	model "github.com/threefoldtech/tfexplorer/models/workloads"
	"github.com/threefoldtech/tfexplorer/mw"
	capacitytypes "github.com/threefoldtech/tfexplorer/pkg/capacity/types"
	phonebook "github.com/threefoldtech/tfexplorer/pkg/phonebook/types"
	"github.com/threefoldtech/tfexplorer/pkg/workloads/types"
	"github.com/threefoldtech/tfexplorer/schema"
)

func (a *API) create(r *http.Request) (interface{}, mw.Response) {
	defer r.Body.Close()

	var codec model.Codec
	if err := json.NewDecoder(r.Body).Decode(&codec); err != nil {
		return nil, mw.BadRequest(err)
	}
	workload := codec.Workloader

	// we make sure those arrays are initialized correctly
	// this will make updating the document in place much easier
	// in later stages
	contract := workload.GetContract()
	contract.ID = schema.ID(0)

	state := workload.GetState()
	state.Result = model.Result{}
	state.SignatureFarmer = model.SigningSignature{}
	state.SignaturesDelete = []model.SigningSignature{}
	state.SignaturesProvision = []model.SigningSignature{}

	if err := validateReservation(workload); err != nil {
		return nil, mw.BadRequest(err)
	}

	var err error
	workload, err = a.workloadpipeline(workload, nil)
	if err != nil {
		// if failed to create pipeline, then
		// this reservation has failed initial validation
		return nil, mw.BadRequest(err)
	}

	if state.IsAny(types.Invalid, types.Delete) {
		return nil, mw.BadRequest(fmt.Errorf("invalid request wrong status '%s'", state.NextAction.String()))
	}

	db := mw.Database(r)

	var filter phonebook.UserFilter
	filter = filter.WithID(schema.ID(contract.CustomerTid))
	user, err := filter.Get(r.Context(), db)
	if err != nil {
		return nil, mw.BadRequest(errors.Wrapf(err, "cannot find user with id '%d'", contract.CustomerTid))
	}

	signature, err := hex.DecodeString(state.CustomerSignature)
	if err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "invalid signature format, expecting hex encoded string"))
	}

	if err := model.Verify(workload, user.Pubkey, signature); err != nil {
		return nil, mw.BadRequest(errors.Wrap(err, "failed to verify customer signature"))
	}

	contract.Epoch = schema.Date{Time: time.Now()}

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
			log.Error().Err(err).Int64("poolID", contract.PoolID).Msg("pool disappeared")
			return nil, mw.Error(errors.New("pool does not exist"))
		}
		log.Error().Err(err).Msg("failed to load workload capacity pool")
		return nil, mw.Error(errors.New("could not load the required capacity pool"))
	}

	if !allowed {
		log.Debug().Msg("don't deploy workload as its pool is almost empty")
		if err := types.WorkloadSetNextAction(r.Context(), db, id, model.NextActionInvalid); err != nil {
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

	ws := []model.Workloader{}
	for cur.Next(r.Context()) {
		var w model.Codec
		if err := cur.Decode(&w); err != nil {
			// skip reservations we can not load
			// this is probably an old reservation
			currentID := cur.Current.Lookup("_id").Int64()
			log.Error().Err(err).Int64("id", currentID).Msg("failed to decode reservation")
			continue
		}

		workload, err := a.workloadpipeline(w.Workloader, nil)
		if err != nil {
			log.Error().Err(err).Int64("id", int64(workload.GetContract().ID)).Msg("failed to process reservation")
			continue
		}

		ws = append(ws, workload)
	}

	pages := fmt.Sprintf("%d", models.NrPages(total, *pager.Limit))
	return ws, mw.Ok().WithHeader("Pages", pages)
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

func (a *API) signProvision(r *http.Request) (interface{}, mw.Response) {
	var signature model.SigningSignature

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

	if reservation.NextAction != model.NextActionSign {
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

	if reservation.NextAction == model.NextActionDeploy {
		types.WorkloadPush(r.Context(), db, reservation.Workloads("")...)
	}

	return nil, mw.Created()
}

func (a *API) newSignProvision(r *http.Request) (interface{}, mw.Response) {
	var signature model.SigningSignature

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

	s := workload.GetState()
	c := workload.GetContract()

	if s.NextAction != model.NextActionSign {
		return nil, mw.UnAuthorized(fmt.Errorf("workload not expecting signatures"))
	}

	if httpErr := userCanSign(signature.Tid, c.SigningRequestProvision, s.SignaturesProvision); httpErr != nil {
		return nil, httpErr
	}

	user, err := phonebook.UserFilter{}.WithID(schema.ID(signature.Tid)).Get(r.Context(), db)
	if err != nil {
		return nil, mw.NotFound(errors.Wrap(err, "customer id not found"))
	}

	if err := model.SignatureProvisionRequestVerify(workload, user.Pubkey, signature); err != nil {
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

	if s.NextAction == model.NextActionDeploy {
		if err = types.WorkloadPush(r.Context(), db, workload); err != nil {
			return nil, mw.Error(err)
		}
	}

	return nil, mw.Created()
}

func (a *API) signDelete(r *http.Request) (interface{}, mw.Response) {
	var signature model.SigningSignature

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

	if reservation.NextAction != model.NextActionDelete {
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
	var signature model.SigningSignature

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

	c := workload.GetContract()

	if httpErr := userCanSign(signature.Tid, c.SigningRequestDelete, workload.GetState().SignaturesDelete); httpErr != nil {
		return nil, httpErr
	}

	user, err := phonebook.UserFilter{}.WithID(schema.ID(signature.Tid)).Get(r.Context(), db)
	if err != nil {
		return nil, mw.NotFound(errors.Wrap(err, "customer id not found"))
	}

	if err := model.SignatureDeleteRequestVerify(workload, user.Pubkey, signature); err != nil {
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

	if workload.GetState().NextAction != model.NextActionDelete {
		return nil, mw.Created()
	}

	_, err = a.setWorkloadDelete(r.Context(), db, workload)
	if err != nil {
		return nil, mw.Error(err)
	}

	return nil, mw.Created()
}
