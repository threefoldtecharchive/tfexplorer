package directory

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/zaibon/httpsig"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfexplorer/models"
	generated "github.com/threefoldtech/tfexplorer/models/generated/directory"
	"github.com/threefoldtech/tfexplorer/mw"
	directory "github.com/threefoldtech/tfexplorer/pkg/directory/types"

	"github.com/gorilla/mux"
)

var (
	errFailedToRegisterGateway = errors.New("failed to register gateway")
	errFailedToListGateways    = errors.New("failed to list gateways")
)

func (s *GatewayAPI) registerGateway(r *http.Request) (interface{}, mw.Response) {
	log.Info().Msg("gateway register request received")

	defer r.Body.Close()

	var gw directory.Gateway
	if err := json.NewDecoder(r.Body).Decode(&gw); err != nil {
		return nil, mw.BadRequest(err)
	}

	db := mw.Database(r)
	if _, err := s.Add(r.Context(), db, gw); err != nil {
		return nil, mw.MongoError(mw.MongoDBError{Cause: err, Message: errFailedToRegisterGateway.Error()})
	}

	log.Info().Msgf("gateway registered: %+v\n", gw)

	return nil, mw.Created()
}

func (s *GatewayAPI) gatewayDetail(r *http.Request) (interface{}, mw.Response) {
	gatewayID := mux.Vars(r)["gateway_id"]
	q := gatewayQuery{}
	if err := q.Parse(r); err != nil {
		return nil, err
	}
	db := mw.Database(r)

	gateway, err := s.Get(r.Context(), db, gatewayID)
	if err != nil {
		return nil, mw.MongoError(mw.MongoDBError{Cause: err, Message: fmt.Sprintf("gateway with id %s not found", gatewayID)})
	}

	return gateway, nil
}

func (s *GatewayAPI) listGateways(r *http.Request) (interface{}, mw.Response) {
	q := gatewayQuery{}
	if err := q.Parse(r); err != nil {
		return nil, mw.Error(errFailedToListGateways)
	}

	db := mw.Database(r)
	pager := models.PageFromRequest(r)
	gateways, total, err := s.List(r.Context(), db, q, pager)
	if err != nil {
		return nil, mw.MongoError(mw.MongoDBError{Cause: err, Message: errFailedToListGateways.Error()})
	}

	pages := fmt.Sprintf("%d", models.Pages(pager, total))
	return gateways, mw.Ok().WithHeader("Pages", pages)
}

func (s *GatewayAPI) updateUptimeHandler(r *http.Request) (interface{}, mw.Response) {
	defer r.Body.Close()

	gatewayID := mux.Vars(r)["gateway_id"]
	hGatewayID := httpsig.KeyIDFromContext(r.Context())
	if gatewayID != hGatewayID {
		return nil, mw.Forbidden(fmt.Errorf("trying to register uptime for gatewayID %s while you are %s", gatewayID, hGatewayID))
	}

	input := struct {
		Uptime uint64 `json:"uptime"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return nil, mw.BadRequest(err)
	}

	db := mw.Database(r)
	log.Debug().Str("gateway", gatewayID).Uint64("uptime", input.Uptime).Msg("gateway uptime received")

	if err := s.updateUptime(r.Context(), db, gatewayID, int64(input.Uptime)); err != nil {
		return nil, mw.MongoError(mw.MongoDBError{Cause: err, Message: fmt.Sprintf("failed to update gateway with gatewayID %s", gatewayID)})
	}

	return nil, nil
}

func (s *GatewayAPI) updateReservedResources(r *http.Request) (interface{}, mw.Response) {
	defer r.Body.Close()

	gatewayID := mux.Vars(r)["gateway_id"]
	hGatewayID := httpsig.KeyIDFromContext(r.Context())
	if gatewayID != hGatewayID {
		return nil, mw.Forbidden(fmt.Errorf("trying to update reserved capacity for gatewayID %s while you are %s", gatewayID, hGatewayID))
	}

	input := struct {
		generated.ResourceAmount
		generated.WorkloadAmount
	}{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return nil, mw.BadRequest(err)
	}

	db := mw.Database(r)
	if err := s.updateReservedCapacity(r.Context(), db, gatewayID, input.ResourceAmount); err != nil {
		return nil, mw.MongoError(mw.MongoDBError{Cause: err, Message: fmt.Sprintf("failed to update gateway reserved capacity with gatewayID %s", gatewayID)})
	}
	if err := s.updateWorkloadsAmount(r.Context(), db, gatewayID, input.WorkloadAmount); err != nil {
		return nil, mw.MongoError(mw.MongoDBError{Cause: err, Message: fmt.Sprintf("failed to update gateway workloads amount with gatewayID %s", gatewayID)})
	}

	return nil, nil
}
