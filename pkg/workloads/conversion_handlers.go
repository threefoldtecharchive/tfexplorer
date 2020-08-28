package workloads

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfexplorer/models/workloads"
	"github.com/threefoldtech/tfexplorer/mw"
	"github.com/threefoldtech/tfexplorer/pkg/capacity"
	capacitytypes "github.com/threefoldtech/tfexplorer/pkg/capacity/types"
	directorytypes "github.com/threefoldtech/tfexplorer/pkg/directory/types"
	phonebooktypes "github.com/threefoldtech/tfexplorer/pkg/phonebook/types"
	"github.com/threefoldtech/tfexplorer/pkg/workloads/types"
	"github.com/threefoldtech/tfexplorer/schema"
	"github.com/zaibon/httpsig"
	"go.mongodb.org/mongo-driver/mongo"
)

func (a *API) getConversionList(r *http.Request) (interface{}, mw.Response) {
	sUserTid := httpsig.KeyIDFromContext(r.Context())
	userTid, err := strconv.ParseInt(sUserTid, 10, 64)
	if err != nil {
		return nil, mw.BadRequest(err)
	}

	db := mw.Database(r)

	// check if we already generated conversion workloads
	cd, err := types.GetUserConversion(r.Context(), db, schema.ID(userTid))
	if err != nil && !errors.Is(err, types.ErrNoConversion) {
		return nil, mw.Error(err)
	}

	if err == nil {
		if cd.Converted {
			// no more data
			return nil, mw.NoContent()
		}

		return cd.Workloads, mw.Ok()
	}

	resPerFarm := map[int64][]workloads.Workloader{}

	// get all reservations for a user
	reservations, err := a.reservationsForUser(r.Context(), db, userTid)
	if err != nil {
		return nil, mw.Error(err)
	}

	workloaders, err := loadWorkloaders(reservations)
	if err != nil {
		return nil, mw.Error(err)
	}

	networks, err := loadNetworks(reservations)
	if err != nil {
		return nil, mw.Error(err)
	}

	workloaders = append(workloaders, networks...)

	for _, w := range workloaders {
		c := w.GetContract()
		nodeID := c.NodeID
		farmID, err := farmForNodeID(r.Context(), db, nodeID)
		if err != nil {
			return nil, mw.Error(err)
		}
		resPerFarm[farmID] = append(resPerFarm[farmID], w)

		// normalize container volumes
		if c.WorkloadType == workloads.WorkloadTypeContainer {
			cont := w.(*workloads.Container)
			for i, vol := range cont.Volumes {
				if strings.HasPrefix(vol.VolumeId, "-") {
					cont.Volumes[i].VolumeId = fmt.Sprintf("%d%s", cont.GetContract().ID, vol.VolumeId)
				}
			}
		}
	}

	for farmID := range resPerFarm {
		nodeIDs, err := farmNodeIDs(r.Context(), db, farmID)
		if err != nil {
			return nil, mw.Error(err)
		}
		// create pool
		pool := capacitytypes.NewPool(0, userTid, nodeIDs)
		pool, err = capacitytypes.CapacityPoolCreate(r.Context(), db, pool)
		if err != nil {
			return nil, mw.Error(err)
		}

		// set pool id on the workloads
		wls := resPerFarm[farmID]
		for i := range wls {
			wls[i].GetContract().PoolID = int64(pool.ID)
		}
		resPerFarm[farmID] = wls
	}

	if err = types.SaveUserConversion(r.Context(), db, schema.ID(userTid), workloaders); err != nil {
		return nil, mw.Error(err)
	}

	return workloaders, mw.Ok()
}

func (a *API) postConversionList(r *http.Request) (interface{}, mw.Response) {
	var workloaders []workloads.Workloader
	if err := json.NewDecoder(r.Body).Decode(&workloaders); err != nil {
		return nil, mw.BadRequest(err)
	}

	if len(workloaders) == 0 {
		return nil, mw.BadRequest(errors.New("need to send at least 1 workload to convert"))
	}

	db := mw.Database(r)

	// load the user. We just pick the first workload to fetch the customer_tid.
	// This works as all workloads need to be owned by the same user
	userTid := workloaders[0].GetContract().CustomerTid

	// make sure user id is the id of the user who signed the request. Later on,
	// we verify that all workloads have the same user, therefore, if this one
	// workload is done by the correct user, all of them are.
	var userFilter phonebooktypes.UserFilter
	userFilter = userFilter.WithID(schema.ID(userTid))
	user, err := userFilter.Get(r.Context(), db)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, mw.BadRequest(errors.New("customer not found"))
		}
		return nil, mw.Error(err)
	}

	// check if we already generated conversion workloads
	cd, err := types.GetUserConversion(r.Context(), db, schema.ID(userTid))
	if err != nil && !errors.Is(err, types.ErrNoConversion) {
		return nil, mw.Error(err)
	}
	if errors.Is(err, types.ErrNoConversion) {
		return nil, mw.NotFound(err)
	}

	if cd.Converted {
		// no more data
		return nil, mw.Conflict(errors.New("conversion already ran"))
	}

	expectedWls := cd.Workloads

	if len(expectedWls) != len(workloaders) {
		return nil, mw.BadRequest(errors.New("unexpected amount of workloads"))
	}

	emptyResult := workloads.Result{}
	for i := range workloaders {
		es := expectedWls[i].GetState()
		ws := workloaders[i].GetState()

		// customer signature should be the only change
		es.CustomerSignature = ws.CustomerSignature

		// we skip the result object when comparing the expected and received workloads
		savedResult := ws.Result //we save it though so we can put it back before saving into the DB
		ws.Result = emptyResult
		es.Result = emptyResult

		// skip signature farmer, it's always empty anyhow
		es.SignatureFarmer = workloads.SigningSignature{}
		ws.SignatureFarmer = workloads.SigningSignature{}

		// truncate time to account for the lost nanosecond precision during json marshalling
		workloaders[i].GetContract().Epoch = schema.Date{Time: workloaders[i].GetContract().Epoch.Time.Truncate(time.Second)}
		expectedWls[i].GetContract().Epoch = schema.Date{Time: expectedWls[i].GetContract().Epoch.Time.Truncate(time.Second)}

		// use string representation cause reflect.DeepEqual was impossible to get right
		expected := fmt.Sprintf("%+v", expectedWls[i])
		received := fmt.Sprintf("%+v", workloaders[i])
		if expected != received {
			return nil, mw.BadRequest(errors.New("invalid workload"))
		}

		sig, err := hex.DecodeString(ws.CustomerSignature)
		if err != nil {
			return nil, mw.BadRequest(err)
		}

		if err = workloads.Verify(workloaders[i], user.Pubkey, sig); err != nil {
			return nil, mw.BadRequest(fmt.Errorf("workload %d (%s) signature verification failed: %w", workloaders[i].GetContract().ID, workloaders[i].GetContract().WorkloadType.String(), err))
		}

		// set the result back in place to be added into the db
		ws.Result = savedResult
	}

	// calculate how much to add per pool
	poolCU := make(map[int64]float64)
	poolSU := make(map[int64]float64)
	for _, wl := range workloaders {
		ss := strings.Split(wl.GetContract().Reference, "-")
		reservationID, err := a.parseID(ss[0])
		if err != nil {
			return nil, mw.Error(err)
		}
		reservation, err := types.ReservationFilter{}.WithID(reservationID).Get(r.Context(), db)
		if err != nil {
			return nil, mw.Error(err)
		}
		if reservation.Expired() {
			// should not happen
			continue
		}
		secondsLeft := math.Floor(time.Until(reservation.DataReservation.ExpirationReservation.Time).Seconds())
		cu, su := capacity.CloudUnitsFromResourceUnits(wl.GetRSU())
		poolID := wl.GetContract().PoolID

		log.Info().Msgf("pool %d cu %v su %v %+v", poolID, cu, su, wl.GetContract().WorkloadType.String())

		if cu > 0 {
			poolCU[poolID] = poolCU[poolID] + cu*secondsLeft
		}
		if su > 0 {
			poolSU[poolID] = poolSU[poolID] + su*secondsLeft
		}
	}

	// this is fine since these pools should not be used yet
	// TODO is it really though
	for poolID := range poolCU {
		pool, err := capacitytypes.GetPool(r.Context(), db, schema.ID(poolID))
		if err != nil {
			return nil, mw.Error(err)
		}
		pool.AddCapacity(poolCU[poolID], poolSU[poolID])
		if err = capacitytypes.UpdatePool(r.Context(), db, pool); err != nil {
			return nil, mw.Error(err)
		}
	}

	// all reservations are as created and have valid signatures
	for i := range workloaders {
		workloaders[i].GetContract().ID = 0 //force to create a new workload ID
		if _, err = types.WorkloadCreate(r.Context(), db, workloaders[i]); err != nil {
			return nil, mw.Error(err)
		}
		if workloaders[i].GetState().Result.State == workloads.ResultStateOK {
			if err := a.capacityPlanner.AddUsedCapacity(workloaders[i]); err != nil {
				return nil, mw.Error(err)
			}
		}

		// Marked the migrated reservation as migrated so it is never send to the node anymore
		ss := strings.Split(workloaders[i].GetContract().Reference, "-")
		rid, err := strconv.Atoi(ss[0])
		if err != nil {
			return nil, mw.Error(err)
		}

		if err := types.ReservationSetNextAction(r.Context(), db, schema.ID(rid), workloads.NextActionMigrated); err != nil {
			return nil, mw.Error(err)
		}
	}

	if err = types.SetUserConversionSucceeded(r.Context(), db, schema.ID(userTid)); err != nil {
		return nil, mw.Error(err)
	}

	return nil, nil
}

func (a *API) reservationsForUser(ctx context.Context, db *mongo.Database, userTid int64) ([]types.Reservation, error) {
	var filter types.ReservationFilter
	filter = filter.WithCustomerID(userTid)

	cur, err := filter.Find(ctx, db)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open reservation cursor")
	}
	defer cur.Close(ctx)

	reservations := []types.Reservation{}

	for cur.Next(ctx) {
		var reservation types.Reservation
		if err := cur.Decode(&reservation); err != nil {
			// skip reservations we can not load
			// this is probably an old reservation
			currentID := cur.Current.Lookup("_id").Int64()
			log.Error().Err(err).Int64("id", currentID).Msg("failed to decode reservation")
			return nil, errors.Wrap(err, "faield to decode reservation")
		}

		reservation, err := a.pipeline(reservation, nil)
		if err != nil {
			log.Error().Err(err).Int64("id", int64(reservation.ID)).Msg("failed to process reservation")
			return nil, errors.Wrap(err, "failed to process reservation")
		}

		reservations = append(reservations, reservation)
	}

	return reservations, nil
}

func loadWorkloaders(res []types.Reservation) ([]workloads.Workloader, error) {
	workloaders := make([]workloads.Workloader, 0, len(res))
	for _, r := range res {
		for _, w := range r.Workloads("") {
			if w.GetContract().WorkloadType == workloads.WorkloadTypeNetwork ||
				w.GetContract().WorkloadType == workloads.WorkloadTypeNetworkResource {
				continue
			}

			if w.GetState().Result.State != workloads.ResultStateOK {
				continue
			}
			w.GetContract().Reference = fmt.Sprintf("%d-%d", r.ID, w.GetContract().WorkloadID)
			workloaders = append(workloaders, w)
		}
	}

	return workloaders, nil
}

func loadNetworks(res []types.Reservation) ([]workloads.Workloader, error) {
	networkReservation := map[string]types.Reservation{}

	for _, r := range res {
		if r.NextAction != workloads.NextActionDeploy {
			continue
		}
		for _, network := range r.DataReservation.Networks {
			networkReservation[network.Name] = r
		}
	}

	workloaders := make([]workloads.Workloader, 0)
	for _, r := range networkReservation {
		for i := range r.DataReservation.Networks {
			network := r.DataReservation.Networks[i]
			networkResources := network.ToNetworkResources()
			for i := range networkResources {
				nr := networkResources[i]
				c := nr.GetContract()
				s := nr.GetState()

				c.CustomerTid = r.CustomerTid
				s.NextAction = r.NextAction
				c.ID = r.ID
				c.Epoch = r.Epoch
				c.Metadata = r.Metadata
				c.Reference = fmt.Sprintf("%d-%d", r.ID, network.WorkloadID())
				c.Description = r.DataReservation.Description
				c.SigningRequestDelete = r.DataReservation.SigningRequestDelete
				c.SigningRequestProvision = r.DataReservation.SigningRequestProvision
				workloaders = append(workloaders, &nr)
				for _, result := range r.Results {
					if result.NodeId == c.NodeID {
						s.Result = result
					}
				}
			}
		}
	}

	return workloaders, nil
}

// farmForNodeID return the farm id in which the node or gateway lives
func farmForNodeID(ctx context.Context, db *mongo.Database, nodeID string) (int64, error) {
	var nodeFilter directorytypes.NodeFilter
	nodeFilter = nodeFilter.WithNodeID(nodeID)
	var gwFilter directorytypes.GatewayFilter
	gwFilter = gwFilter.WithGWID(nodeID)
	var farmID int64

	node, err := nodeFilter.Get(ctx, db, false)
	if err == nil {
		farmID = node.FarmId
	} else {
		gw, err := gwFilter.Get(ctx, db)
		if err != nil {
			return 0, err
		}
		farmID = gw.FarmId
	}

	return int64(farmID), nil
}

func farmNodeIDs(ctx context.Context, db *mongo.Database, farmID int64) ([]string, error) {
	var filter directorytypes.NodeFilter
	filter = filter.WithFarmID(schema.ID(farmID))

	var nodes []directorytypes.Node
	cur, err := filter.Find(ctx, db)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	if err := cur.All(ctx, &nodes); err != nil {
		return nil, err
	}

	nodesID := make([]string, len(nodes))
	for i, node := range nodes {
		nodesID[i] = node.NodeId
	}

	return nodesID, nil
}
