package directory

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/zaibon/httpsig"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/threefoldtech/tfexplorer/models"
	generated "github.com/threefoldtech/tfexplorer/models/generated/directory"
	"github.com/threefoldtech/tfexplorer/mw"
	"github.com/threefoldtech/tfexplorer/pkg/directory/types"
	directory "github.com/threefoldtech/tfexplorer/pkg/directory/types"
	"github.com/threefoldtech/tfexplorer/schema"

	"github.com/gorilla/mux"
)

func (s *GatewayAPI) registerGateway(r *http.Request) (interface{}, mw.Response) {
	log.Info().Msg("node register request received")

	defer r.Body.Close()

	var gw directory.Gateway
	if err := json.NewDecoder(r.Body).Decode(&gw); err != nil {
		return nil, mw.BadRequest(err)
	}

	keyID := httpsig.KeyIDFromContext(r.Context())
	if gw.NodeId != keyID {
		return nil, mw.Forbidden(fmt.Errorf("trying to register a gateway with nodeID %s while you are %s", gw.NodeId, keyID))
	}

	db := mw.Database(r)

	var ff types.FarmFilter
	ff = ff.WithID(schema.ID(gw.FarmId))
	_, err := ff.Get(r.Context(), db)
	if err != mongo.ErrNoDocuments {
		return nil, mw.NotFound(errors.Wrap(err, "farm not found"))
	} else if err != nil {
		return nil, mw.Error(err)
	}

	if err := gw.Validate(); err != nil {
		return nil, mw.BadRequest(err)
	}

	for _, domain := range gw.ManagedDomains {
		if err := s.isManagedDomain(gw.NodeId, domain); err != nil {
			return nil, mw.Forbidden(err)
		}
	}

	if _, err := s.Add(r.Context(), db, gw); err != nil {
		return nil, mw.Error(err)
	}

	log.Info().Msgf("gateway registered: %+v\n", gw)

	return nil, mw.Created()
}

func (s *GatewayAPI) isManagedDomain(identity, domain string) error {
	const name = "__owner__"
	host := fmt.Sprintf("%s.%s", name, domain)

	records, err := net.LookupTXT(host)
	if err != nil {
		return errors.Wrapf(err, "failed to look up '%s' for TXT records", host)
	}

	var value struct {
		Identity string `json:"identity"`
		Owner    string `json:"owner"`
	}

	for _, record := range records {
		if err := json.Unmarshal([]byte(record), &value); err != nil {
			log.Error().Err(err).Str("host", host).Msg("txt record for host is not valid json")
			continue
		}

		// record found
		// we match both the identity (validate this domain is managed by the gateway)
		// and also the owner (validate that this way is owned by the gateway not another user)
		if value.Identity == identity &&
			value.Owner == identity {
			return nil
		}
	}

	return fmt.Errorf("failed to validate managed domain '%s'. no txt recrod with identity '%s' found", domain, identity)
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
