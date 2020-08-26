package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/stellar/go/support/errors"
	"github.com/threefoldtech/tfexplorer/models/workloads"
	"github.com/threefoldtech/tfexplorer/pkg/capacity/types"
	wrklds "github.com/threefoldtech/tfexplorer/pkg/workloads"
	wrkldstypes "github.com/threefoldtech/tfexplorer/pkg/workloads/types"
	"github.com/threefoldtech/tfexplorer/schema"
)

type httpWorkloads struct {
	*httpClient
}

func (w *httpWorkloads) Create(workload workloads.Workloader) (resp wrklds.ReservationCreateResponse, err error) {
	_, err = w.post(w.url("workloads"), workload, &resp, http.StatusCreated)
	return
}

func (w *httpWorkloads) List(nextAction *workloads.NextActionEnum, customerTid int64, page *Pager) (reservations []workloads.Reservation, err error) {
	query := url.Values{}
	if nextAction != nil {
		query.Set("next_action", fmt.Sprintf("%d", *nextAction))
	}
	if customerTid != 0 {
		query.Set("customer_tid", fmt.Sprint(customerTid))
	}
	page.apply(query)

	_, err = w.get(w.url("workloads"), query, &reservations, http.StatusOK)
	return
}

func (w *httpWorkloads) Get(id schema.ID) (workload workloads.Workloader, err error) {
	_, err = w.get(w.url("workloads", fmt.Sprint(id)), nil, &workload, http.StatusOK)
	return
}

func (w *httpWorkloads) SignProvision(id schema.ID, user schema.ID, signature string) error {
	_, err := w.post(
		w.url("workloads", fmt.Sprint(id), "sign", "provision"),
		workloads.SigningSignature{
			Tid:       int64(user),
			Signature: signature,
		},
		nil,
		http.StatusCreated,
	)

	return err
}

func (w *httpWorkloads) SignDelete(id schema.ID, user schema.ID, signature string) error {
	_, err := w.post(
		w.url("workloads", fmt.Sprint(id), "sign", "delete"),
		workloads.SigningSignature{
			Tid:       int64(user),
			Signature: signature,
		},
		nil,
		http.StatusCreated,
	)

	return err
}

func (w *httpWorkloads) NodeWorkloads(nodeID string, from uint64) ([]workloads.Workloader, uint64, error) {
	query := url.Values{}
	query.Set("from", fmt.Sprint(from))

	var list []wrkldstypes.WorkloaderType

	u := w.url("reservations", "nodes", nodeID, "workloads")
	if len(query) > 0 {
		u = fmt.Sprintf("%s?%s", u, query.Encode())
	}

	response, err := http.Get(u)
	if err != nil {
		return nil, 0, err
	}

	if err := w.process(response, &list, http.StatusOK); err != nil {
		return nil, 0, err
	}

	var lastID uint64
	if idStr := response.Header.Get("x-last-id"); len(idStr) != 0 {
		lastID, err = strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return nil, lastID, errors.Wrap(err, "failed to extract last id value")
		}
	}

	output := make([]workloads.Workloader, len(list))
	for i, w := range list {
		output[i] = w.Workloader
	}

	return output, lastID, err
}

func (w *httpWorkloads) NodeWorkloadGet(gwid string) (result workloads.Workloader, err error) {
	// var output intermediateWL
	var output workloads.Workloader
	_, err = w.get(w.url("reservations", "nodes", "workloads", gwid), nil, &output, http.StatusOK)
	if err != nil {
		return
	}

	return output, nil
}

func (w *httpWorkloads) NodeWorkloadPutResult(nodeID, gwid string, result workloads.Result) error {
	_, err := w.put(w.url("reservations", "nodes", nodeID, "workloads", gwid), result, nil, http.StatusCreated)
	return err
}

func (w *httpWorkloads) NodeWorkloadPutDeleted(nodeID, gwid string) error {
	_, err := w.delete(w.url("reservations", "nodes", nodeID, "workloads", gwid), nil, nil, http.StatusOK)
	return err
}

func (w *httpWorkloads) PoolCreate(reservation types.Reservation) (resp wrklds.CapacityPoolCreateResponse, err error) {
	_, err = w.post(w.url("reservations", "pools"), reservation, &resp, http.StatusCreated)
	return
}

func (w *httpWorkloads) PoolGet(poolID string) (result types.Pool, err error) {
	var pool types.Pool
	_, err = w.get(w.url("reservations", "pools", poolID), nil, &pool, http.StatusOK)
	if err != nil {
		return
	}

	return pool, nil
}

func (w *httpWorkloads) PoolsGetByOwner(ownerID string) (result []types.Pool, err error) {
	var pools []types.Pool
	_, err = w.get(w.url("reservations", "pools", "owner", ownerID), nil, &pools, http.StatusOK)
	if err != nil {
		return
	}

	return pools, nil
}
