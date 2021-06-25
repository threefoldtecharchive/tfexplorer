package workloads

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/threefoldtech/tfexplorer/mw"
	"github.com/threefoldtech/tfexplorer/pkg/capacity"
	"github.com/threefoldtech/tfexplorer/pkg/escrow"
	"github.com/threefoldtech/tfexplorer/pkg/gridnetworks"
	"github.com/threefoldtech/tfexplorer/pkg/workloads/types"
	"github.com/zaibon/httpsig"
	"go.mongodb.org/mongo-driver/mongo"
)

// Setup injects and initializes directory package
func Setup(parent *mux.Router, db *mongo.Database, network gridnetworks.GridNetwork, escrow escrow.Escrow, planner capacity.Planner) error {
	if err := types.Setup(context.TODO(), db); err != nil {
		return err
	}

	userVerifier := httpsig.NewVerifier(mw.NewUserKeyGetter(db))

	service := API{
		escrow:          escrow,
		capacityPlanner: planner,
		network:         network,
	}

	// versionned endpoints
	api := parent.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/prices", mw.AsHandlerFunc(service.getPrices)).Methods(http.MethodGet).Name("prices-get")

	apiReservation := api.PathPrefix("/reservations").Subrouter()

	apiReservation.HandleFunc("/pools", mw.AsHandlerFunc(service.setupPool)).Methods(http.MethodPost).Name("versionned-pool-create")
	apiReservation.HandleFunc("/pools/{id:\\d+}", mw.AsHandlerFunc(service.getPool)).Methods(http.MethodGet).Name("versionned-pool-get")
	apiReservation.HandleFunc("/pools/owner/{owner:\\d+}", mw.AsHandlerFunc(service.listPools)).Methods(http.MethodGet).Name("versionned-pool-get-by-owner")
	apiReservation.HandleFunc("/pools/payment/{id:\\d+}", mw.AsHandlerFunc(service.getPaymentInfo)).Methods(http.MethodGet).Name("versionned-pool-get-payment-info")
	// only create reservation call requires authentication to make sure
	// the user identity associated with the request is the same exact
	// one associated with the signed reservation object.
	authenticated := apiReservation.NewRoute().Subrouter()
	authenticated.Use(mw.NewAuthMiddleware(userVerifier).Middleware)
	authenticated.HandleFunc("/workloads", mw.AsHandlerFunc(service.create)).Methods(http.MethodPost).Name("versionned-workloads-create")
	// other calls are public
	apiReservation.HandleFunc("/workloads", mw.AsHandlerFunc(service.listWorkload)).Methods(http.MethodGet).Name("versionned-workloadreservation-list")
	apiReservation.HandleFunc("/workloads/{res_id:\\d+}", mw.AsHandlerFunc(service.getWorkload)).Methods(http.MethodGet).Name("versionned-workloadreservation-get")
	apiReservation.HandleFunc("/workloads/{res_id:\\d+}/sign/provision", mw.AsHandlerFunc(service.signProvision)).Methods(http.MethodPost).Name("versionned-reservation-sign-provision")
	apiReservation.HandleFunc("/workloads/{res_id:\\d+}/sign/delete", mw.AsHandlerFunc(service.newSignDelete)).Methods(http.MethodPost).Name("versionned-reservation-sign-delete")

	conversionAuthenticated := apiReservation.PathPrefix("/convert").Subrouter()
	conversionAuthenticated.Use(mw.NewAuthMiddleware(userVerifier).Middleware)
	conversionAuthenticated.HandleFunc("", mw.AsHandlerFunc(service.getConversionList)).Methods(http.MethodGet).Name("versionned-conversion-list")
	conversionAuthenticated.HandleFunc("", mw.AsHandlerFunc(service.postConversionList)).Methods(http.MethodPost).Name("versionned-conversion-post")

	// Nodes oriented endpoints
	apiReservation.HandleFunc("/nodes/{node_id}/workloads", mw.AsHandlerFunc(service.workloads)).Queries("from", "{from:\\d+}").Methods(http.MethodGet).Name("versionned-workloads-poll")
	apiReservation.HandleFunc("/nodes/workloads/{gwid:\\d+-\\d+}", mw.AsHandlerFunc(service.workloadGet)).Methods(http.MethodGet).Name("versionned-workload-get")
	apiReservation.HandleFunc("/nodes/{node_id}/workloads/{gwid:\\d+-\\d+}", mw.AsHandlerFunc(service.workloadPutResult)).Methods(http.MethodPut).Name("versionned-workloads-results")
	apiReservation.HandleFunc("/nodes/{node_id}/workloads/{gwid:\\d+-\\d+}", mw.AsHandlerFunc(service.workloadPutDeleted)).Methods(http.MethodDelete).Name("versionned-workloads-deleted")
	apiReservation.HandleFunc("/nodes/{node_id}/workloads/{gwid:\\d+-\\d+}", mw.AsHandlerFunc(service.workloadPutConsumption)).Methods(http.MethodPatch).Name("versionned-workloads-update-consumption")

	// legacy endpoints
	legacyReservations := parent.PathPrefix("/explorer/reservations").Subrouter()

	legacyReservations.HandleFunc("", mw.AsHandlerFunc(service.create)).Methods(http.MethodPost).Name("reservation-create")
	legacyReservations.HandleFunc("", mw.AsHandlerFunc(service.list)).Methods(http.MethodGet).Name("reservation-list")
	legacyReservations.HandleFunc("/{res_id:\\d+}", mw.AsHandlerFunc(service.get)).Methods(http.MethodGet).Name("reservation-get")
	legacyReservations.HandleFunc("/{res_id:\\d+}/sign/provision", mw.AsHandlerFunc(service.signProvision)).Methods(http.MethodPost).Name("reservation-sign-provision")
	legacyReservations.HandleFunc("/{res_id:\\d+}/sign/delete", mw.AsHandlerFunc(service.signDelete)).Methods(http.MethodPost).Name("reservation-sign-delete")

	// new style workloads
	workloads := parent.PathPrefix("/explorer/workloads").Subrouter()
	workloads.HandleFunc("", mw.AsHandlerFunc(service.create)).Methods(http.MethodPost).Name("workload-create")
	workloads.HandleFunc("", mw.AsHandlerFunc(service.listWorkload)).Methods(http.MethodGet).Name("workload-list")
	workloads.HandleFunc("/{res_id:\\d+}", mw.AsHandlerFunc(service.getWorkload)).Methods(http.MethodGet).Name("workload-get")
	workloads.HandleFunc("/{res_id:\\d+}/sign/provision", mw.AsHandlerFunc(service.signProvision)).Methods(http.MethodPost).Name("workload-sign-provision")
	workloads.HandleFunc("/{res_id:\\d+}/sign/delete", mw.AsHandlerFunc(service.signDelete)).Methods(http.MethodPost).Name("workload-sign-delete")

	legacyReservations.HandleFunc("/pools", mw.AsHandlerFunc(service.setupPool)).Methods(http.MethodPost).Name("pool-create")
	legacyReservations.HandleFunc("/pools/{id:\\d+}", mw.AsHandlerFunc(service.getPool)).Methods(http.MethodGet).Name("pool-get")
	legacyReservations.HandleFunc("/pools/owner/{owner:\\d+}", mw.AsHandlerFunc(service.listPools)).Methods(http.MethodGet).Name("pool-get-by-owner")

	// conversion
	legacyConversionAuthenticated := legacyReservations.PathPrefix("/explorer/convert").Subrouter()
	legacyConversionAuthenticated.Use(mw.NewAuthMiddleware(userVerifier).Middleware)
	legacyConversionAuthenticated.HandleFunc("", mw.AsHandlerFunc(service.getConversionList)).Methods(http.MethodGet).Name("conversion-list")
	legacyConversionAuthenticated.HandleFunc("", mw.AsHandlerFunc(service.postConversionList)).Methods(http.MethodPost).Name("conversion-post")

	// node oriented endpoints
	legacyReservations.HandleFunc("/nodes/{node_id}/workloads", mw.AsHandlerFunc(service.workloads)).Queries("from", "{from:\\d+}").Methods(http.MethodGet).Name("nodes-workloads-poll")
	legacyReservations.HandleFunc("/workloads/{gwid:\\d+-\\d+}", mw.AsHandlerFunc(service.workloadGet)).Methods(http.MethodGet).Name("nodes-workload-get")
	legacyReservations.HandleFunc("/nodes/{node_id}/workloads/{gwid:\\d+-\\d+}", mw.AsHandlerFunc(service.workloadPutResult)).Methods(http.MethodPut).Name("nodes-workloads-results")
	legacyReservations.HandleFunc("/nodes/{node_id}/workloads/{gwid:\\d+-\\d+}", mw.AsHandlerFunc(service.workloadPutDeleted)).Methods(http.MethodDelete).Name("nodes-workloads-deleted")

	return nil
}
