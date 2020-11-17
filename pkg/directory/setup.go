package directory

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/threefoldtech/tfexplorer/mw"
	directory "github.com/threefoldtech/tfexplorer/pkg/directory/types"
	"github.com/zaibon/httpsig"
	"go.mongodb.org/mongo-driver/mongo"
)

// Setup injects and initializes directory package
func Setup(parent *mux.Router, db *mongo.Database) error {
	if err := directory.Setup(context.TODO(), db); err != nil {
		return err
	}

	userVerifier := httpsig.NewVerifier(mw.NewUserKeyGetter(db))
	nodeVerifier := httpsig.NewVerifier(mw.NewNodeKeyGetter())

	var farmAPI = FarmAPI{
		verifier: userVerifier,
	}
	var nodeAPI NodeAPI

	// versionned endpoints
	api := parent.PathPrefix("/api/v1").Subrouter()
	farms := api.PathPrefix("/farms").Subrouter()

	farms.HandleFunc("", mw.AsHandlerFunc(farmAPI.registerFarm)).Methods("POST").Name("farm-register-v1")
	farms.HandleFunc("", mw.AsHandlerFunc(farmAPI.listFarm)).Methods("GET").Name("farm-list-v1")
	farms.HandleFunc("/{farm_id}", mw.AsHandlerFunc(farmAPI.getFarm)).Methods("GET").Name("farm-get-v1")

	farmsAuthenticated := api.PathPrefix("/farms").Subrouter()
	farmsAuthenticated.Use(mw.NewAuthMiddleware(userVerifier).Middleware)
	farmsAuthenticated.HandleFunc("/{farm_id}/ip", farmAPI.verifySameFarm(mw.AsHandlerFunc(farmAPI.addFarmIPs))).Methods("POST").Name("farm-add-ip-v1")
	farmsAuthenticated.HandleFunc("/{farm_id}/ip", farmAPI.verifySameFarm(mw.AsHandlerFunc(farmAPI.deleteFarmIps))).Methods("DELETE").Name("farm-delete-ip-v1")
	farmsAuthenticated.HandleFunc("/{farm_id}", farmAPI.verifySameFarm(mw.AsHandlerFunc(farmAPI.updateFarm))).Methods("PUT").Name("farm-update-v1")
	farmsAuthenticated.HandleFunc("/{farm_id}/{node_id}", farmAPI.verifySameFarm(mw.AsHandlerFunc(nodeAPI.Requires("node_id", farmAPI.deleteNodeFromFarm)))).Methods("DELETE").Name("farm-node-delete-v1")

	nodes := api.PathPrefix("/nodes").Subrouter()
	nodesAuthenticated := api.PathPrefix("/nodes").Subrouter()
	userAuthenticated := api.PathPrefix("/nodes").Subrouter()

	nodeAuthMW := mw.NewAuthMiddleware(nodeVerifier)
	userAuthMW := mw.NewAuthMiddleware(userVerifier)

	userAuthenticated.Use(userAuthMW.Middleware)
	nodesAuthenticated.Use(nodeAuthMW.Middleware)

	nodesAuthenticated.HandleFunc("", mw.AsHandlerFunc(nodeAPI.registerNode)).Methods("POST").Name("node-register-v1")
	nodes.HandleFunc("", mw.AsHandlerFunc(nodeAPI.listNodes)).Methods("GET").Name("nodes-list-v1")
	nodes.HandleFunc("/{node_id}", mw.AsHandlerFunc(nodeAPI.nodeDetail)).Methods("GET").Name(("node-get-v1"))
	nodesAuthenticated.HandleFunc("/{node_id}/interfaces", mw.AsHandlerFunc(nodeAPI.Requires("node_id", nodeAPI.registerIfaces))).Methods("POST").Name("node-interfaces-v1")
	nodesAuthenticated.HandleFunc("/{node_id}/ports", mw.AsHandlerFunc(nodeAPI.Requires("node_id", nodeAPI.registerPorts))).Methods("POST").Name("node-set-ports-v1")
	userAuthenticated.HandleFunc("/{node_id}/configure_public", mw.AsHandlerFunc(nodeAPI.Requires("node_id", nodeAPI.configurePublic))).Methods("POST").Name("node-configure-public-v1")
	userAuthenticated.HandleFunc("/{node_id}/configure_free", mw.AsHandlerFunc(nodeAPI.Requires("node_id", nodeAPI.configureFreeToUse))).Methods("POST").Name("node-configure-free-v1")
	nodesAuthenticated.HandleFunc("/{node_id}/capacity", mw.AsHandlerFunc(nodeAPI.Requires("node_id", nodeAPI.registerCapacity))).Methods("POST").Name("node-capacity-v1")
	nodesAuthenticated.HandleFunc("/{node_id}/uptime", mw.AsHandlerFunc(nodeAPI.Requires("node_id", nodeAPI.updateUptimeHandler))).Methods("POST").Name("node-uptime-v1")
	nodesAuthenticated.HandleFunc("/{node_id}/used_resources", mw.AsHandlerFunc(nodeAPI.Requires("node_id", nodeAPI.updateReservedResources))).Methods("POST").Name("node-reserved-resources-v1")

	var gwAPI GatewayAPI
	gw := api.PathPrefix("/gateways").Subrouter()
	gwAuthenticated := api.PathPrefix("/gateways").Subrouter()
	gwAuthMW := mw.NewAuthMiddleware(nodeVerifier)
	gwAuthenticated.Use(gwAuthMW.Middleware)

	gw.HandleFunc("", mw.AsHandlerFunc(gwAPI.registerGateway)).Methods("POST").Name("gateway-register-v1")
	gw.HandleFunc("", mw.AsHandlerFunc(gwAPI.listGateways)).Methods("GET").Name("gateway-list-v1")
	gw.HandleFunc("/{node_id}", mw.AsHandlerFunc(gwAPI.gatewayDetail)).Methods("GET").Name(("gateway-get-v1"))
	gwAuthenticated.HandleFunc("/{node_id}/uptime", mw.AsHandlerFunc(gwAPI.Requires("node_id", gwAPI.updateUptimeHandler))).Methods("POST").Name("gateway-uptime-v1")
	gwAuthenticated.HandleFunc("/{node_id}/reserved_resources", mw.AsHandlerFunc(gwAPI.Requires("node_id", gwAPI.updateReservedResources))).Methods("POST").Name("gateway-reserved-resources-v1")

	// legacy endpoints
	legacyFarms := parent.PathPrefix("/explorer/farms").Subrouter()
	legacyFarmsAuthenticated := parent.PathPrefix("/explorer/farms").Subrouter()
	legacyFarmsAuthenticated.Use(mw.NewAuthMiddleware(userVerifier).Middleware)

	legacyFarms.HandleFunc("", mw.AsHandlerFunc(farmAPI.registerFarm)).Methods("POST").Name("farm-register")
	legacyFarms.HandleFunc("", mw.AsHandlerFunc(farmAPI.listFarm)).Methods("GET").Name("farm-list")
	legacyFarms.HandleFunc("/{farm_id}", mw.AsHandlerFunc(farmAPI.getFarm)).Methods("GET").Name("farm-get")
	legacyFarmsAuthenticated.HandleFunc("/{farm_id}", mw.AsHandlerFunc(farmAPI.updateFarm)).Methods("PUT").Name("farm-update")
	legacyFarmsAuthenticated.HandleFunc("/{farm_id}/{node_id}", mw.AsHandlerFunc(nodeAPI.Requires("node_id", farmAPI.deleteNodeFromFarm))).Methods("DELETE").Name("farm-node-delete")

	legacyNodes := parent.PathPrefix("/explorer/nodes").Subrouter()
	legacyNodesAuthenticated := parent.PathPrefix("/explorer/nodes").Subrouter()
	legacyUserAuthenticated := parent.PathPrefix("/explorer/nodes").Subrouter()

	legacyNodeAuthMW := mw.NewAuthMiddleware(nodeVerifier)
	legacyUserAuthMW := mw.NewAuthMiddleware(userVerifier)

	legacyUserAuthenticated.Use(legacyUserAuthMW.Middleware)
	legacyNodesAuthenticated.Use(legacyNodeAuthMW.Middleware)

	legacyNodesAuthenticated.HandleFunc("", mw.AsHandlerFunc(nodeAPI.registerNode)).Methods("POST").Name("node-register")
	legacyNodes.HandleFunc("", mw.AsHandlerFunc(nodeAPI.listNodes)).Methods("GET").Name("nodes-list")
	legacyNodes.HandleFunc("/{node_id}", mw.AsHandlerFunc(nodeAPI.nodeDetail)).Methods("GET").Name(("node-get"))
	legacyNodesAuthenticated.HandleFunc("/{node_id}/interfaces", mw.AsHandlerFunc(nodeAPI.Requires("node_id", nodeAPI.registerIfaces))).Methods("POST").Name("node-interfaces")
	legacyNodesAuthenticated.HandleFunc("/{node_id}/ports", mw.AsHandlerFunc(nodeAPI.Requires("node_id", nodeAPI.registerPorts))).Methods("POST").Name("node-set-ports")
	legacyUserAuthenticated.HandleFunc("/{node_id}/configure_public", mw.AsHandlerFunc(nodeAPI.Requires("node_id", nodeAPI.configurePublic))).Methods("POST").Name("node-configure-public")
	legacyUserAuthenticated.HandleFunc("/{node_id}/configure_free", mw.AsHandlerFunc(nodeAPI.Requires("node_id", nodeAPI.configureFreeToUse))).Methods("POST").Name("node-configure-free")
	legacyNodesAuthenticated.HandleFunc("/{node_id}/capacity", mw.AsHandlerFunc(nodeAPI.Requires("node_id", nodeAPI.registerCapacity))).Methods("POST").Name("node-capacity")
	legacyNodesAuthenticated.HandleFunc("/{node_id}/uptime", mw.AsHandlerFunc(nodeAPI.Requires("node_id", nodeAPI.updateUptimeHandler))).Methods("POST").Name("node-uptime")
	legacyNodesAuthenticated.HandleFunc("/{node_id}/used_resources", mw.AsHandlerFunc(nodeAPI.Requires("node_id", nodeAPI.updateReservedResources))).Methods("POST").Name("node-reserved-resources")

	legacyGw := parent.PathPrefix("/explorer/gateways").Subrouter()
	legacyGwAuthenticated := parent.PathPrefix("/explorer/gateways").Subrouter()
	legacyGwAuthMW := mw.NewAuthMiddleware(nodeVerifier)
	legacyGwAuthenticated.Use(legacyGwAuthMW.Middleware)

	legacyGw.HandleFunc("", mw.AsHandlerFunc(gwAPI.registerGateway)).Methods("POST").Name("gateway-register")
	legacyGw.HandleFunc("", mw.AsHandlerFunc(gwAPI.listGateways)).Methods("GET").Name("gateway-list")
	legacyGw.HandleFunc("/{node_id}", mw.AsHandlerFunc(gwAPI.gatewayDetail)).Methods("GET").Name(("gateway-get"))
	legacyGwAuthenticated.HandleFunc("/{node_id}/uptime", mw.AsHandlerFunc(gwAPI.Requires("node_id", gwAPI.updateUptimeHandler))).Methods("POST").Name("gateway-uptime")
	legacyGwAuthenticated.HandleFunc("/{node_id}/reserved_resources", mw.AsHandlerFunc(gwAPI.Requires("node_id", gwAPI.updateReservedResources))).Methods("POST").Name("gateway-reserved-resources")

	return nil
}
