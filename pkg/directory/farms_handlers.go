package directory

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/zaibon/httpsig"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/threefoldtech/tfexplorer/models"
	"github.com/threefoldtech/tfexplorer/mw"
	directory "github.com/threefoldtech/tfexplorer/pkg/directory/types"
	"github.com/threefoldtech/tfexplorer/pkg/events"
	"github.com/threefoldtech/tfexplorer/schema"

	"github.com/gorilla/mux"
)

func (f FarmAPI) isAuthenticated(r *http.Request) bool {
	_, err := f.verifier.Verify(r)
	return err == nil
}

type myWriter struct {
	http.Hijacker
	http.ResponseWriter
}

func (f *FarmAPI) wsEnpoint(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%+v", w)
	events.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// upgrade this connection to a WebSocket connection
	ws, err := events.Upgrader.Upgrade(&myWriter{ResponseWriter: w}, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to upgrade connection")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error occured"))
		return
	}
	for {
		f.eventProcessingService.SendEvents(ws)
	}
}

func (f *FarmAPI) registerFarm(r *http.Request) (interface{}, mw.Response) {
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

	id, err := f.Add(r.Context(), db, info)
	if err != nil {
		return nil, mw.Error(err)
	}

	return struct {
		ID schema.ID `json:"id"`
	}{
		id,
	}, mw.Created()
}

func (f *FarmAPI) updateFarm(r *http.Request) (interface{}, mw.Response) {
	sid := mux.Vars(r)["farm_id"]

	id, err := strconv.ParseInt(sid, 10, 64)
	if err != nil {
		return nil, mw.BadRequest(err)
	}

	db := mw.Database(r)

	farm, err := f.GetByID(r.Context(), db, id)
	if err != nil {
		return nil, mw.NotFound(err)
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

	err = f.Update(r.Context(), db, info.ID, info)
	if err != nil {
		return nil, mw.Error(err)
	}

	return nil, mw.Ok()
}

func (f *FarmAPI) listFarm(r *http.Request) (interface{}, mw.Response) {

	q := directory.FarmQuery{}
	if err := q.Parse(r); err != nil {
		return nil, err
	}
	var filter directory.FarmFilter
	filter = filter.WithFarmQuery(q)

	db := mw.Database(r)

	var findOpts []*options.FindOptions

	pager := models.PageFromRequest(r)
	findOpts = append(findOpts, pager)

	// hide the email of the farm for any non authenticated user
	if !f.isAuthenticated(r) {
		findOpts = append(findOpts, options.Find().SetProjection(bson.D{
			{Key: "email", Value: 0},
		}))
	}

	farms, total, err := f.List(r.Context(), db, filter, findOpts...)
	if err != nil {
		return nil, mw.Error(err)
	}

	pages := fmt.Sprintf("%d", models.Pages(pager, total))
	return farms, mw.Ok().WithHeader("Pages", pages)
}

func (f *FarmAPI) getFarm(r *http.Request) (interface{}, mw.Response) {
	sid := mux.Vars(r)["farm_id"]

	id, err := strconv.ParseInt(sid, 10, 64)
	if err != nil {
		return nil, mw.BadRequest(err)
	}

	db := mw.Database(r)

	farm, err := f.GetByID(r.Context(), db, id)
	if err != nil {
		return nil, mw.NotFound(err)
	}

	// hide the email of the farm for any non authenticated user
	if !f.isAuthenticated(r) {
		farm.Email = ""
	}

	return farm, nil
}

func (f *FarmAPI) deleteNodeFromFarm(r *http.Request) (interface{}, mw.Response) {
	sid := mux.Vars(r)["farm_id"]

	id, err := strconv.ParseInt(sid, 10, 64)
	if err != nil {
		return nil, mw.BadRequest(err)
	}

	db := mw.Database(r)

	farm, err := f.GetByID(r.Context(), db, id)
	if err != nil {
		return nil, mw.NotFound(err)
	}

	sfarmerID := httpsig.KeyIDFromContext(r.Context())
	requestFarmerID, err := strconv.ParseInt(sfarmerID, 10, 64)
	if err != nil {
		return nil, mw.BadRequest(err)
	}

	if farm.ThreebotId != requestFarmerID {
		return nil, mw.Forbidden(fmt.Errorf("only the farm owner of this node can delete this node"))
	}

	var nodeAPI NodeAPI
	nodeID := mux.Vars(r)["node_id"]

	node, err := nodeAPI.Get(r.Context(), db, nodeID, false)
	if err != nil {
		return nil, mw.NotFound(err)
	}

	if node.FarmId != int64(farm.ID) {
		return nil, mw.Forbidden(fmt.Errorf("only the farm owner of this node can delete this node"))
	}

	err = nodeAPI.Delete(r.Context(), db, nodeID)
	if err != nil {
		return nil, mw.Error(err)
	}

	return nil, mw.Ok()
}
