package workloads

import schema "github.com/threefoldtech/tfexplorer/schema"

func (n NetworkResource) WorkloadID() int64 {
	return n.WorkloadId
}

type NetworkResource struct {
	Name                         string            `bson:"name" json:"name"`
	WorkloadId                   int64             `bson:"workload_id" json:"workload_id"`
	NodeId                       string            `bson:"node_id" json:"node_id"`
	StatsAggregator              []StatsAggregator `bson:"stats_aggregator" json:"stats_aggregator"`
	WireguardPrivateKeyEncrypted string            `bson:"wireguard_private_key_encrypted" json:"wireguard_private_key_encrypted"`
	WireguardPublicKey           string            `bson:"wireguard_public_key" json:"wireguard_public_key"`
	WireguardListenPort          int64             `bson:"wireguard_listen_port" json:"wireguard_listen_port"`
	Iprange                      schema.IPRange    `bson:"iprange" json:"iprange"`
	Peers                        []WireguardPeer   `bson:"peers" json:"peers"`
	PoolId                       int64             `bson:"pool_id" json:"pool_id"`

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
	Results             []Result           `bson:"results" json:"results"`
}
