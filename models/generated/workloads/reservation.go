package workloads

import (
	"encoding/json"

	schema "github.com/threefoldtech/tfexplorer/schema"
)

type Reservation struct {
	ID                  schema.ID          `bson:"_id" json:"id"`
	Json                string             `bson:"json" json:"json"`
	DataReservation     ReservationData    `bson:"data_reservation" json:"data_reservation"`
	CustomerTid         int64              `bson:"customer_tid" json:"customer_tid"`
	CustomerSignature   string             `bson:"customer_signature" json:"customer_signature"`
	NextAction          NextActionEnum     `bson:"next_action" json:"next_action"`
	SignaturesProvision []SigningSignature `bson:"signatures_provision" json:"signatures_provision"`
	SignaturesFarmer    []SigningSignature `bson:"signatures_farmer" json:"signatures_farmer"`
	SignaturesDelete    []SigningSignature `bson:"signatures_delete" json:"signatures_delete"`
	Epoch               schema.Date        `bson:"epoch" json:"epoch"`
	Metadata            string             `bson:"metadata" json:"metadata"`
	Results             []Result           `bson:"results" json:"results"`
}

type NextActionEnum uint8

const (
	NextActionCreate NextActionEnum = iota
	NextActionSign
	NextActionPay
	NextActionDeploy
	NextActionDelete
	NextActionInvalid
	NextActionDeleted
	NextActionMigrated
)

func (e NextActionEnum) String() string {
	switch e {
	case NextActionCreate:
		return "create"
	case NextActionSign:
		return "sign"
	case NextActionPay:
		return "pay"
	case NextActionDeploy:
		return "deploy"
	case NextActionDelete:
		return "delete"
	case NextActionInvalid:
		return "invalid"
	case NextActionDeleted:
		return "deleted"
	}
	return "UNKNOWN"
}

type ReservationData struct {
	Description             string                `bson:"description" json:"description"`
	Currencies              []string              `bson:"currencies" json:"currencies"`
	SigningRequestProvision SigningRequest        `bson:"signing_request_provision" json:"signing_request_provision"`
	SigningRequestDelete    SigningRequest        `bson:"signing_request_delete" json:"signing_request_delete"`
	Containers              []Container           `bson:"containers" json:"containers"`
	Volumes                 []Volume              `bson:"volumes" json:"volumes"`
	Zdbs                    []ZDB                 `bson:"zdbs" json:"zdbs"`
	Networks                []Network             `bson:"networks" json:"networks"`
	NetworkResources        []NetworkResource     `bson:"network_resource" json:"network_resource"`
	Kubernetes              []K8S                 `bson:"kubernetes" json:"kubernetes"`
	Proxies                 []GatewayProxy        `bson:"proxies" json:"proxies"`
	ReverseProxy            []GatewayReverseProxy `bson:"reverse_proxies" json:"reverse_proxies"`
	Subdomains              []GatewaySubdomain    `bson:"subdomains" json:"subdomains"`
	DomainDelegates         []GatewayDelegate     `bson:"domain_delegates" json:"domain_delegates"`
	Gateway4To6s            []Gateway4To6         `bson:"gateway4to6" json:"gateway4to6"`
	ExpirationProvisioning  schema.Date           `bson:"expiration_provisioning" json:"expiration_provisioning"`
	ExpirationReservation   schema.Date           `bson:"expiration_reservation" json:"expiration_reservation"`
}

type SigningRequest struct {
	Signers   []int64 `bson:"signers" json:"signers"`
	QuorumMin int64   `bson:"quorum_min" json:"quorum_min"`
}

type SigningSignature struct {
	Tid       int64       `bson:"tid" json:"tid"`
	Signature string      `bson:"signature" json:"signature"`
	Epoch     schema.Date `bson:"epoch" json:"epoch"`
}

type Result struct {
	Category   WorkloadTypeEnum `bson:"category" json:"category"`
	WorkloadId string           `bson:"workload_id" json:"workload_id"`
	DataJson   json.RawMessage  `bson:"data_json" json:"data_json"`
	Signature  string           `bson:"signature" json:"signature"`
	State      ResultStateEnum  `bson:"state" json:"state"`
	Message    string           `bson:"message" json:"message"`
	Epoch      schema.Date      `bson:"epoch" json:"epoch"`
	NodeId     string           `bson:"node_id" json:"node_id"`
}

type ResultStateEnum uint8

const (
	ResultStateError ResultStateEnum = iota
	ResultStateOK
	ResultStateDeleted
)

func (e ResultStateEnum) String() string {
	switch e {
	case ResultStateError:
		return "error"
	case ResultStateOK:
		return "ok"
	case ResultStateDeleted:
		return "deleted"
	}
	return "UNKNOWN"
}

type ReservationWorkload struct {
	WorkloadId string           `bson:"workload_id" json:"workload_id"`
	User       string           `bson:"user" json:"user"`
	PoolID     int64            `bson:"pool_id" json:"pool_id"`
	Type       WorkloadTypeEnum `bson:"type" json:"type"`
	Content    interface{}      `bson:"content" json:"content"`
	Created    schema.Date      `bson:"created" json:"created"`
	Duration   int64            `bson:"duration" json:"duration"`
	Signature  string           `bson:"signature" json:"signature"`
	ToDelete   bool             `bson:"to_delete" json:"to_delete"`
}

type WorkloadTypeEnum uint8

const (
	WorkloadTypeZDB WorkloadTypeEnum = iota
	WorkloadTypeContainer
	WorkloadTypeVolume
	WorkloadTypeNetwork
	WorkloadTypeKubernetes
	WorkloadTypeProxy
	WorkloadTypeReverseProxy
	WorkloadTypeSubDomain
	WorkloadTypeDomainDelegate
	WorkloadTypeGateway4To6
	WorkloadTypeNetworkResource
)

// WorkloadTypes is a map of all the supported workload type
var WorkloadTypes = map[WorkloadTypeEnum]string{
	WorkloadTypeZDB:             "zdb",
	WorkloadTypeContainer:       "container",
	WorkloadTypeVolume:          "volume",
	WorkloadTypeNetwork:         "network",
	WorkloadTypeKubernetes:      "kubernetes",
	WorkloadTypeProxy:           "proxy",
	WorkloadTypeReverseProxy:    "reverse_proxy",
	WorkloadTypeSubDomain:       "subdomain",
	WorkloadTypeDomainDelegate:  "domain_delegate",
	WorkloadTypeGateway4To6:     "gateway4to6",
	WorkloadTypeNetworkResource: "network_resource",
}

func (e WorkloadTypeEnum) String() string {
	s, ok := WorkloadTypes[e]
	if !ok {
		return "UNKNOWN"
	}
	return s
}
