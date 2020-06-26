package workloads

import (
	"net"

	schema "github.com/threefoldtech/tfexplorer/schema"
)

type K8S struct {
	WorkloadId      int64             `bson:"workload_id" json:"workload_id"`
	NodeId          string            `bson:"node_id" json:"node_id"`
	Size            int64             `bson:"size" json:"size"`
	NetworkId       string            `bson:"network_id" json:"network_id"`
	Ipaddress       net.IP            `bson:"ipaddress" json:"ipaddress"`
	ClusterSecret   string            `bson:"cluster_secret" json:"cluster_secret"`
	MasterIps       []net.IP          `bson:"master_ips" json:"master_ips"`
	SshKeys         []string          `bson:"ssh_keys" json:"ssh_keys"`
	StatsAggregator []StatsAggregator `bson:"stats_aggregator" json:"stats_aggregator"`
	FarmerTid       int64             `bson:"farmer_tid" json:"farmer_tid"`
	PoolId          int64             `bson:"pool_id" json:"pool_id"`

	Description             string         `bson:"description" json:"description"`
	Currencies              []string       `bson:"currencies" json:"currencies"`
	SigningRequestProvision SigningRequest `bson:"signing_request_provision" json:"signing_request_provision"`
	SigningRequestDelete    SigningRequest `bson:"signing_request_delete" json:"signing_request_delete"`
	ExpirationProvisioning  schema.Date    `bson:"expiration_provisioning" json:"expiration_provisioning"`

	ID                  schema.ID          `bson:"_id" json:"id"`
	Json                string             `bson:"json" json:"json"`
	CustomerTid         int64              `bson:"customer_tid" json:"customer_tid"`
	CustomerSignature   string             `bson:"customer_signature" json:"customer_signature"`
	NextAction          NextActionEnum     `bson:"next_action" json:"next_action"`
	SignaturesProvision []SigningSignature `bson:"signatures_provision" json:"signatures_provision"`
	SignaturesFarmer    []SigningSignature `bson:"signatures_farmer" json:"signatures_farmer"`
	SignaturesDelete    []SigningSignature `bson:"signatures_delete" json:"signatures_delete"`
	Epoch               schema.Date        `bson:"epoch" json:"epoch"`
	Metadata            string             `bson:"metadata" json:"metadata"`
	Result              Result             `bson:"result" json:"result"`
	WorkloadType        WorkloadTypeEnum   `bson:"workload_type" json:"workload_type"`
}

func (k K8S) WorkloadID() int64 {
	return k.WorkloadId

