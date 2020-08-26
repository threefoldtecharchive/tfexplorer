package workloads

import (
	"fmt"
	"io"

	schema "github.com/threefoldtech/tfexplorer/schema"
)

type Network struct {
	Name             string               `bson:"name" json:"name"`
	WorkloadId       int64                `bson:"workload_id" json:"workload_id"`
	Iprange          schema.IPRange       `bson:"iprange" json:"iprange"`
	StatsAggregator  []StatsAggregator    `bson:"stats_aggregator" json:"stats_aggregator"`
	NetworkResources []NetworkNetResource `bson:"network_resources" json:"network_resources"`
	FarmerTid        int64                `bson:"farmer_tid" json:"farmer_tid"`
}

func (n Network) WorkloadID() int64 {
	return n.WorkloadId
}

func (n Network) ToNetworkResources() []NetworkResource {
	netRes := make([]NetworkResource, len(n.NetworkResources))

	for i := range n.NetworkResources {
		nr := NetworkResource{
			contract: Contract{
				WorkloadID:   n.WorkloadID(),
				WorkloadType: WorkloadTypeNetworkResource,
				NodeID:       n.NetworkResources[i].NodeId,
			},
			Name:                         n.Name,
			NetworkIprange:               n.Iprange,
			WireguardPrivateKeyEncrypted: n.NetworkResources[i].WireguardPrivateKeyEncrypted,
			WireguardPublicKey:           n.NetworkResources[i].WireguardPublicKey,
			WireguardListenPort:          n.NetworkResources[i].WireguardListenPort,
			Iprange:                      n.NetworkResources[i].Iprange,
			Peers:                        n.NetworkResources[i].Peers,
		}
		netRes[i] = nr
	}

	return netRes
}

type NetworkNetResource struct {
	NodeId                       string          `bson:"node_id" json:"node_id"`
	WireguardPrivateKeyEncrypted string          `bson:"wireguard_private_key_encrypted" json:"wireguard_private_key_encrypted"`
	WireguardPublicKey           string          `bson:"wireguard_public_key" json:"wireguard_public_key"`
	WireguardListenPort          int64           `bson:"wireguard_listen_port" json:"wireguard_listen_port"`
	Iprange                      schema.IPRange  `bson:"iprange" json:"iprange"`
	Peers                        []WireguardPeer `bson:"peers" json:"peers"`
}

type WireguardPeer struct {
	PublicKey      string           `bson:"public_key" json:"public_key"`
	Endpoint       string           `bson:"endpoint" json:"endpoint"`
	Iprange        schema.IPRange   `bson:"iprange" json:"iprange"`
	AllowedIprange []schema.IPRange `bson:"allowed_iprange" json:"allowed_iprange"`
}

func (p *WireguardPeer) SigingEncode(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "%s", p.PublicKey); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%s", p.Endpoint); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "%s", p.Iprange.String()); err != nil {
		return err
	}
	for _, iprange := range p.AllowedIprange {
		if _, err := fmt.Fprintf(w, "%s", iprange.String()); err != nil {
			return err
		}
	}
	return nil
}
