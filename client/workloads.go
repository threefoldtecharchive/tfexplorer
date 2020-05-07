package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/stellar/go/support/errors"
	"github.com/threefoldtech/tfexplorer/models/generated/workloads"
	wrklds "github.com/threefoldtech/tfexplorer/pkg/workloads"
	"github.com/threefoldtech/tfexplorer/schema"
)

type httpWorkloads struct {
	*httpClient
}

func (w *httpWorkloads) Create(reservation workloads.Reservation) (resp wrklds.ReservationCreateResponse, err error) {
	_, err = w.post(w.url("reservations"), reservation, &resp, http.StatusCreated)
	return
}

func (w *httpWorkloads) List(nextAction *workloads.NextActionEnum, customerTid int64, page *Pager) (reservation []workloads.Reservation, err error) {
	query := url.Values{}
	if nextAction != nil {
		query.Set("next_action", fmt.Sprintf("%d", nextAction))
	}
	if customerTid != 0 {
		query.Set("customer_tid", fmt.Sprint(customerTid))
	}
	page.apply(query)

	_, err = w.get(w.url("reservations"), query, &reservation, http.StatusOK)
	return
}

func (w *httpWorkloads) Get(id schema.ID) (reservation workloads.Reservation, err error) {
	_, err = w.get(w.url("reservations", fmt.Sprint(id)), nil, &reservation, http.StatusOK)
	return
}

func (w *httpWorkloads) SignProvision(id schema.ID, user schema.ID, signature string) error {
	_, err := w.post(
		w.url("reservations", fmt.Sprint(id), "sign", "provision"),
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
		w.url("reservations", fmt.Sprint(id), "sign", "delete"),
		workloads.SigningSignature{
			Tid:       int64(user),
			Signature: signature,
		},
		nil,
		http.StatusCreated,
	)

	return err
}

type intermediateWL struct {
	workloads.ReservationWorkload
	Content json.RawMessage `json:"content"`
}

func (wl *intermediateWL) Workload() (result workloads.ReservationWorkload, err error) {
	result = wl.ReservationWorkload
	switch wl.Type {
	case workloads.WorkloadTypeContainer:
		var o workloads.Container
		if err := json.Unmarshal(wl.Content, &o); err != nil {
			return result, err
		}
		result.Content = o
	case workloads.WorkloadTypeKubernetes:
		var o workloads.K8S
		if err := json.Unmarshal(wl.Content, &o); err != nil {
			return result, err
		}
		result.Content = o
	case workloads.WorkloadTypeNetwork:
		var o workloads.Network
		if err := json.Unmarshal(wl.Content, &o); err != nil {
			return result, err
		}
		result.Content = o
	case workloads.WorkloadTypeVolume:
		var o workloads.Volume
		if err := json.Unmarshal(wl.Content, &o); err != nil {
			return result, err
		}
		result.Content = o
	case workloads.WorkloadTypeZDB:
		var o workloads.ZDB
		if err := json.Unmarshal(wl.Content, &o); err != nil {
			return result, err
		}
		result.Content = o
	case workloads.WorkloadTypeProxy:
		var o workloads.GatewayProxy
		if err := json.Unmarshal(wl.Content, &o); err != nil {
			return result, err
		}
		result.Content = o
	case workloads.WorkloadTypeReverseProxy:
		var o workloads.GatewayReserveProxy
		if err := json.Unmarshal(wl.Content, &o); err != nil {
			return result, err
		}
		result.Content = o
	case workloads.WorkloadTypeSubDomain:
		var o workloads.GatewaySubdomain
		if err := json.Unmarshal(wl.Content, &o); err != nil {
			return result, err
		}
		result.Content = o
	case workloads.WorkloadTypeDomainDelegate:
		var o workloads.GatewayDelegate
		if err := json.Unmarshal(wl.Content, &o); err != nil {
			return result, err
		}
		result.Content = o
	case workloads.WorkloadTypeGateway4To6:
		var o workloads.Gateway4To6
		if err := json.Unmarshal(wl.Content, &o); err != nil {
			return result, err
		}
		result.Content = o
	case workloads.WorkloadTypeQemu:
		var o workloads.Qemu
		if err := json.Unmarshal(wl.Content, &o); err != nil {
			return result, err
		}
		result.Content = o
	default:
		return result, fmt.Errorf("unknown workload type")
	}

	return
}

func (w *httpWorkloads) Workloads(nodeID string, from uint64) ([]workloads.ReservationWorkload, uint64, error) {
	query := url.Values{}
	query.Set("from", fmt.Sprint(from))

	var list []intermediateWL

	response, err := w.get(
		w.url("reservations", "workloads", nodeID),
		query,
		&list,
		http.StatusOK,
	)

	if err != nil {
		return nil, 0, err
	}

	var lastID uint64
	if idStr := response.Header.Get("x-last-id"); len(idStr) != 0 {
		lastID, err = strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return nil, lastID, errors.Wrap(err, "failed to extract last id value")
		}
	}

	results := make([]workloads.ReservationWorkload, 0, len(list))
	for _, i := range list {
		wl, err := i.Workload()
		if err != nil {
			return nil, lastID, err
		}
		results = append(results, wl)
	}

	return results, lastID, err
}

func (w *httpWorkloads) WorkloadGet(gwid string) (result workloads.ReservationWorkload, err error) {
	var output intermediateWL
	_, err = w.get(w.url("reservations", "workloads", gwid), nil, &output, http.StatusOK)
	if err != nil {
		return
	}

	return output.Workload()
}

func (w *httpWorkloads) WorkloadPutResult(nodeID, gwid string, result workloads.Result) error {
	_, err := w.put(w.url("reservations", "workloads", gwid, nodeID), result, nil, http.StatusCreated)
	return err
}

func (w *httpWorkloads) WorkloadPutDeleted(nodeID, gwid string) error {
	_, err := w.delete(w.url("reservations", "workloads", gwid, nodeID), nil, nil, http.StatusOK)
	return err
}
