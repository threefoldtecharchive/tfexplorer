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
}
