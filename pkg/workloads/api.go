package workloads

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfexplorer/models/workloads"
	"github.com/threefoldtech/tfexplorer/mw"
	"github.com/threefoldtech/tfexplorer/pkg/capacity"
	directory "github.com/threefoldtech/tfexplorer/pkg/directory/types"
	"github.com/threefoldtech/tfexplorer/pkg/escrow"
	escrowtypes "github.com/threefoldtech/tfexplorer/pkg/escrow/types"
	"github.com/threefoldtech/tfexplorer/pkg/workloads/types"
	"github.com/threefoldtech/tfexplorer/schema"
	"go.mongodb.org/mongo-driver/mongo"
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

func (a *API) workloadpipeline(w workloads.Workloader, err error) (workloads.Workloader, error) {
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

func (a *API) setReservationDeleted(ctx context.Context, db *mongo.Database, id schema.ID) error {
	// cancel reservation escrow in case the reservation has not yet been deployed
	a.escrow.ReservationCanceled(id)
	// No longer set the reservation as deleted. This means a workload which managed
	// to deploy will stay allive. This code path should not happen (it can only
	// happen just after the upgrade, for reservations with a pending escrow), and
	// its not worth the hassle to manually figure out where to send the tokens.
	return nil
}

func (a *API) setWorkloadDelete(ctx context.Context, db *mongo.Database, w workloads.Workloader) (workloads.Workloader, error) {
	w.State().NextAction = types.Delete

	if err := types.ReservationSetNextAction(ctx, db, w.Contract().ID, types.Delete); err != nil {
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
