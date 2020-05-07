package directory

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/zaibon/httpsig"

	"github.com/threefoldtech/tfexplorer/models"
	"github.com/threefoldtech/tfexplorer/mw"
	directory "github.com/threefoldtech/tfexplorer/pkg/directory/types"
	"github.com/threefoldtech/tfexplorer/schema"

	"github.com/gorilla/mux"
)

var (
	errFailedToRegisterFarm = errors.New("failed to register farm")
	errFailedToListFarms    = errors.New("failed to list farms")
)

func (s *FarmAPI) registerFarm(r *http.Request) (interface{}, mw.Response) {
	log.Info().Msg("farm register request received")

	db := mw.Database(r)
	defer r.Body.Close()

	var info directory.Farm
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		return nil, mw.BadRequest(err)
	}

	if err := info.Validate(); err != nil {
		return nil, mw.BadRequest(err)
	}

	id, err := s.Add(r.Context(), db, info)
	if err != nil {
		return nil, mw.MongoError(mw.MongoDBError{Cause: err, Message: errFailedToRegisterFarm.Error()})
	}

	return struct {
		ID schema.ID `json:"id"`
	}{
		id,
	}, mw.Created()
}

func (s *FarmAPI) updateFarm(r *http.Request) (interface{}, mw.Response) {
	sid := mux.Vars(r)["farm_id"]

	id, err := strconv.ParseInt(sid, 10, 64)
	if err != nil {
		return nil, mw.BadRequest(err)
	}

	db := mw.Database(r)

	farm, err := s.GetByID(r.Context(), db, id)
	if err != nil {
		return nil, mw.MongoError(mw.MongoDBError{Cause: err, Message: fmt.Sprintf("farm with id %s not found", sid)})
	}

	sfarmerID := httpsig.KeyIDFromContext(r.Context())
	requestFarmerID, err := strconv.ParseInt(sfarmerID, 10, 64)
	if err != nil {
		return nil, mw.BadRequest(err)
	}

	if farm.ThreebotId != requestFarmerID {
		return nil, mw.Forbidden(fmt.Errorf("only the farm owner can update the information of its farm"))
	}

	var info directory.Farm
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		return nil, mw.BadRequest(err)
	}

	info.ID = schema.ID(id)

	err = s.Update(r.Context(), db, info.ID, info)
	if err != nil {
		return nil, mw.MongoError(mw.MongoDBError{Cause: err, Message: fmt.Sprintf("failed to update farm with farmID %s", sid)})
	}

	return nil, mw.Ok()
}

func (s *FarmAPI) listFarm(r *http.Request) (interface{}, mw.Response) {
	q := directory.FarmQuery{}
	if err := q.Parse(r); err != nil {
		return nil, err
	}
	var filter directory.FarmFilter
	filter = filter.WithFarmQuery(q)
	db := mw.Database(r)

	pager := models.PageFromRequest(r)
	farms, total, err := s.List(r.Context(), db, filter, pager)
	if err != nil {
		return nil, mw.MongoError(mw.MongoDBError{Cause: err, Message: errFailedToListFarms.Error()})
	}

	pages := fmt.Sprintf("%d", models.Pages(pager, total))
	return farms, mw.Ok().WithHeader("Pages", pages)
}

func (s *FarmAPI) getFarm(r *http.Request) (interface{}, mw.Response) {
	sid := mux.Vars(r)["farm_id"]

	id, err := strconv.ParseInt(sid, 10, 64)
	if err != nil {
		return nil, mw.BadRequest(err)
	}

	db := mw.Database(r)

	farm, err := s.GetByID(r.Context(), db, id)
	if err != nil {
		return nil, mw.MongoError(mw.MongoDBError{Cause: err, Message: fmt.Sprintf("farm with id %s not found", sid)})
	}

	return farm, nil
}
