package directory

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/zaibon/httpsig"

	"github.com/threefoldtech/tfexplorer/models"
	generated "github.com/threefoldtech/tfexplorer/models/generated/directory"
	"github.com/threefoldtech/tfexplorer/mw"
	directory "github.com/threefoldtech/tfexplorer/pkg/directory/types"

	"github.com/gorilla/mux"
)

func (s *GatewayAPI) registerGateway(r *http.Request) (interface{}, mw.Response) {
	log.Info().Msg("node register request received")

	defer r.Body.Close()

	var gw directory.Gateway
	if err := json.NewDecoder(r.Body).Decode(&gw); err != nil {
		return nil, mw.BadRequest(err)
	}

	db := mw.Database(r)
	if _, err := s.Add(r.Context(), db, gw); err != nil {
		return nil, mw.Error(err)
	}

	log.Info().Msgf("gateway registered: %+v\n", gw)

	return nil, mw.Created()
}

func (s *GatewayAPI) gatewayDetail(r *http.Request) (interface{}, mw.Response) {
	nodeID := mux.Vars(r)["node_id"]
	q := gatewayQuery{}
	if err := q.Parse(r); err != nil {
		return nil, err
	}
	db := mw.Database(r)

	node, err := s.Get(r.Context(), db, nodeID)
	if err != nil {
		return nil, mw.NotFound(err)
	}

	return node, nil
}

func (s *GatewayAPI) listGateways(r *http.Request) (interface{}, mw.Response) {
	q := gatewayQuery{}
	if err := q.Parse(r); err != nil {
		return nil, err
	}

	db := mw.Database(r)
	pager := models.PageFromRequest(r)
	nodes, total, err := s.List(r.Context(), db, q, pager)
	if err != nil {
		return nil, mw.Error(err)
	}

	pages := fmt.Sprintf("%d", models.Pages(pager, total))
	return nodes, mw.Ok().WithHeader("Pages", pages)
}

func (s *GatewayAPI) updateUptimeHandler(r *http.Request) (interface{}, mw.Response) {
	defer r.Body.Close()

	nodeID := mux.Vars(r)["node_id"]
	hNodeID := httpsig.KeyIDFromContext(r.Context())
	if nodeID != hNodeID {
		return nil, mw.Forbidden(fmt.Errorf("trying to register uptime for nodeID %s while you are %s", nodeID, hNodeID))
	}

	input := struct {
		Uptime uint64 `json:"uptime"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return nil, mw.BadRequest(err)
	}

	db := mw.Database(r)
	log.Debug().Str("gateway", nodeID).Uint64("uptime", input.Uptime).Msg("gateway uptime received")

	if err := s.updateUptime(r.Context(), db, nodeID, int64(input.Uptime)); err != nil {
		return nil, mw.NotFound(err)
	}

	return nil, nil
}

func (s *GatewayAPI) updateReservedResources(r *http.Request) (interface{}, mw.Response) {
	defer r.Body.Close()

	nodeID := mux.Vars(r)["node_id"]
	hNodeID := httpsig.KeyIDFromContext(r.Context())
	if nodeID != hNodeID {
		return nil, mw.Forbidden(fmt.Errorf("trying to update reserved capacity for nodeID %s while you are %s", nodeID, hNodeID))
	}

	input := struct {
		generated.ResourceAmount
		generated.WorkloadAmount
	}{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return nil, mw.BadRequest(err)
	}

	db := mw.Database(r)
	if err := s.updateReservedCapacity(r.Context(), db, nodeID, input.ResourceAmount); err != nil {
		return nil, mw.NotFound(err)
	}
	if err := s.updateWorkloadsAmount(r.Context(), db, nodeID, input.WorkloadAmount); err != nil {
		return nil, mw.NotFound(err)
	}

	return nil, nil
}
